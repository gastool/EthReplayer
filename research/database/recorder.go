package database

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/research/bbolt"
	"github.com/ethereum/go-ethereum/research/model"
	"github.com/ethereum/go-ethereum/rlp"
	"sync"
	"sync/atomic"
)

var saveTask = int64(0)

var Redeploy = &sync.Map{}

func wait() {
	for {
		if atomic.LoadInt64(&saveTask) < 10*int64(MaxBatchSize) {
			break
		}
	}
}

func InitRedeploy() {
	DataBases[CodeChange].View(func(tx *bbolt.Tx) error {
		cur := tx.Cursor()
		for k0, _ := cur.First(); k0 != nil; k0, _ = cur.Next() {
			c := tx.Bucket(k0).Cursor()
			redeployIndex := uint32(0)
			for k, v := c.First(); k != nil; k, v = c.Next() {
				var change model.CodeChange
				err := rlp.DecodeBytes(v, &change)
				if err != nil {
					panic(err)
				}
				if change.Redeploy {
					redeployIndex++
				}
			}
			Redeploy.Store(common.BytesToHash(k0), redeployIndex)
		}
		return nil
	})
}

func SaveTxInfo(info *model.TxInfo, index *model.BtIndex) {
	if ReplayMode {
		return
	}
	value, err := rlp.EncodeToBytes(info)
	if err != nil {
		panic(err)
	}
	key := index.ToByte()
	wait()
	atomic.AddInt64(&saveTask, 1)
	go func() {
		defer atomic.AddInt64(&saveTask, -1)
		err = DataBases[Info].Batch(func(tx *bbolt.Tx) error {
			c := tx.Bucket([]byte("tx"))
			return c.Put(key, value)
		})
		if err != nil {
			panic(err)
		}
	}()
}

func SaveBlockInfo(info *model.BlockInfo, index model.BtIndex) {
	if ReplayMode {
		return
	}
	value, err := rlp.EncodeToBytes(info)
	if err != nil {
		panic(err)
	}
	key := index.BlockToByte()
	wait()
	atomic.AddInt64(&saveTask, 1)
	go func() {
		defer atomic.AddInt64(&saveTask, -1)
		e := DataBases[Info].Batch(func(tx *bbolt.Tx) error {
			c := tx.Bucket([]byte("block"))
			return c.Put(key, value)
		})
		if e != nil {
			panic(err)
		}
	}()
}

func SaveAccountState(state *model.AccountState, addr common.Address, index model.BtIndex) {
	if ReplayMode {
		return
	}
	value, err := rlp.EncodeToBytes(state)
	if err != nil {
		panic(err)
	}
	key := index.ToSortKey(nil)
	wait()
	atomic.AddInt64(&saveTask, 1)
	go func() {
		defer atomic.AddInt64(&saveTask, -1)
		err = DataBases[Account].Batch(func(tx *bbolt.Tx) error {
			c, err := tx.CreateBucketIfNotExists(addr.Bytes())
			if err != nil {
				return err
			}
			return c.Put(key, value)
		})
		if err != nil {
			panic(err)
		}
	}()
}

func SaveCode(code []byte, codeHash []byte, addrHash common.Hash, bt model.BtIndex) {
	if ReplayMode {
		if Debug {
			log.Info("debug code", "code", hex.EncodeToString(code),
				"codeHash", hex.EncodeToString(code), "addrHash", addrHash)
		}
		return
	}
	if len(code) == 0 {
		return
	}
	wait()
	atomic.AddInt64(&saveTask, 1)
	go func() {
		defer atomic.AddInt64(&saveTask, -1)
		err := DataBases[Code].Batch(func(tx *bbolt.Tx) error {
			c := tx.Bucket([]byte("code"))
			return c.Put(codeHash[:], code)
		})
		if err != nil {
			panic(err)
		}
	}()
	if v, ok := Redeploy.Load(addrHash); ok {
		change := &model.CodeChange{
			Delete:   false,
			Redeploy: true,
		}
		ch, _ := rlp.EncodeToBytes(change)
		wait()
		atomic.AddInt64(&saveTask, 1)
		key := bt.ToByte()
		go func() {
			defer atomic.AddInt64(&saveTask, -1)
			DataBases[CodeChange].Batch(func(tx *bbolt.Tx) error {
				c := tx.Bucket(addrHash.Bytes())
				return c.Put(key, ch)
			})
		}()
		vv := v.(uint32) + 1
		Redeploy.Store(addrHash, vv)
	}
}

func Suicide(addrHash common.Hash, bt model.BtIndex) {
	if ReplayMode {
		if Debug {
			log.Info("debug suicide", "addrHash", addrHash)
		}
		return
	}
	change := &model.CodeChange{
		Delete:   true,
		Redeploy: false,
	}
	ch, _ := rlp.EncodeToBytes(change)
	key := bt.ToByte()
	wait()
	atomic.AddInt64(&saveTask, 1)
	go func() {
		defer atomic.AddInt64(&saveTask, -1)
		DataBases[CodeChange].Batch(func(tx *bbolt.Tx) error {
			c, _ := tx.CreateBucketIfNotExists(addrHash.Bytes())
			return c.Put(key, ch)
		})
	}()
	if _, ok := Redeploy.Load(addrHash); !ok {
		Redeploy.Store(addrHash, uint32(0))
	}
}

func SaveStorage(storageChange map[common.Hash]common.Hash, addrHash common.Hash, index model.BtIndex) {
	if ReplayMode || len(storageChange) == 0 {
		return
	}
	redeployCount := uint32(0)
	if v, ok := Redeploy.Load(addrHash); ok {
		redeployCount = v.(uint32)
	}
	wait()
	atomic.AddInt64(&saveTask, 1)
	go func() {
		defer atomic.AddInt64(&saveTask, -1)
		err := DataBases[Storage].Batch(func(tx *bbolt.Tx) error {
			storageAddr := addrHash
			if redeployCount > 0 {
				bs := make([]byte, 4)
				binary.LittleEndian.PutUint32(bs, redeployCount)
				storageAddr = crypto.Keccak256Hash(addrHash[:], bs)
			}
			b2, err := tx.CreateBucketIfNotExists(storageAddr[:])
			if err != nil {
				return err
			}
			for k, v := range storageChange {
				key := common.TrimLeftZeroes(k[:])
				if len(key) == 0 {
					key = []byte{0}
				}
				value := common.TrimLeftZeroes(v[:])
				b3, err := b2.CreateBucketIfNotExists(key)
				if err != nil {
					return err
				}
				sk := index.ToSortKey(nil)
				value2 := make([]byte, len(value))
				copy(value2, value)
				err = b3.Put(sk, value2)
				if err != nil {
					return err
				}
			}
			return err
		})
		if err != nil {
			panic(err)
		}
	}()
}

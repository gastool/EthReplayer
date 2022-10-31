package database

import (
	"encoding/binary"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/research/bbolt"
	"github.com/ethereum/go-ethereum/research/model"
	"github.com/ethereum/go-ethereum/rlp"
	"sync"
	"sync/atomic"
)

var saveTask = int64(0)

var Redeploy = &sync.Map{}

func InitRedeploy() {
	DataBases[CodeChange].View(func(tx *bbolt.Tx) error {
		cur := tx.Cursor()
		for k0, _ := cur.First(); k0 != nil; k0, _ = cur.Next() {
			c := tx.Bucket(k0)
			redeployIndex := uint32(0)
			for k, v := c.Cursor().First(); k != nil; k, v = c.Cursor().Next() {
				var change model.CodeChange
				rlp.DecodeBytes(v, &change)
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
	go func() {
		atomic.AddInt64(&saveTask, 1)
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
	go func() {
		atomic.AddInt64(&saveTask, 1)
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
	go func() {
		atomic.AddInt64(&saveTask, 1)
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
	if ReplayMode || len(code) == 0 {
		return
	}
	go func() {
		atomic.AddInt64(&saveTask, 1)
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
		go func() {
			atomic.AddInt64(&saveTask, 1)
			defer atomic.AddInt64(&saveTask, -1)
			DataBases[CodeChange].Batch(func(tx *bbolt.Tx) error {
				c := tx.Bucket(addrHash.Bytes())
				return c.Put(bt.ToByte(), ch)
			})
		}()
		vv := v.(uint32) + 1
		Redeploy.Store(addrHash, vv)
	}
}

func Suicide(addrHash common.Hash, bt model.BtIndex) {
	if ReplayMode {
		return
	}
	change := &model.CodeChange{
		Delete:   true,
		Redeploy: false,
	}
	ch, _ := rlp.EncodeToBytes(change)
	go func() {
		atomic.AddInt64(&saveTask, 1)
		defer atomic.AddInt64(&saveTask, -1)
		DataBases[CodeChange].Batch(func(tx *bbolt.Tx) error {
			c, _ := tx.CreateBucketIfNotExists(addrHash.Bytes())
			return c.Put(bt.ToByte(), ch)
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
	go func() {
		atomic.AddInt64(&saveTask, 1)
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

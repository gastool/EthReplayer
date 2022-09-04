package database

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/research/model"
	"github.com/ethereum/go-ethereum/rlp"
	bolt "go.etcd.io/bbolt"
	"sync/atomic"
)

var parallelTask = int64(0)

func SaveTxInfo(info *model.TxInfo, index *model.BtIndex) {
	value, err := rlp.EncodeToBytes(info)
	if err != nil {
		panic(err)
	}
	key := index.AllToByte()
	go func() {
		atomic.AddInt64(&parallelTask, 1)
		defer atomic.AddInt64(&parallelTask, -1)
		err = DataBases[Info].Batch(func(tx *bolt.Tx) error {
			c, err := tx.CreateBucketIfNotExists([]byte("tx"))
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

func SaveBlockInfo(info *model.BlockInfo, index model.BtIndex) {
	value, err := rlp.EncodeToBytes(info)
	if err != nil {
		panic(err)
	}
	key := index.BlockToByte()
	go func() {
		atomic.AddInt64(&parallelTask, 1)
		defer atomic.AddInt64(&parallelTask, -1)
		e := DataBases[Info].Batch(func(tx *bolt.Tx) error {
			c, err := tx.CreateBucketIfNotExists([]byte("block"))
			if err != nil {
				return err
			}
			return c.Put(key, value)
		})
		if e != nil {
			panic(err)
		}
	}()
}

func SaveAccountState(state *model.AccountState, addr common.Address, index model.BtIndex) {
	value, err := rlp.EncodeToBytes(state)
	if err != nil {
		panic(err)
	}
	key := index.ToSortKey(nil)
	go func() {
		atomic.AddInt64(&parallelTask, 1)
		defer atomic.AddInt64(&parallelTask, -1)
		err = DataBases[Account].Batch(func(tx *bolt.Tx) error {
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

func SaveCode(code []byte, codeHash []byte) {
	go func() {
		atomic.AddInt64(&parallelTask, 1)
		defer atomic.AddInt64(&parallelTask, -1)
		err := DataBases[Code].Update(func(tx *bolt.Tx) error {
			c, err := tx.CreateBucketIfNotExists(codeHash)
			if err != nil {
				return err
			}
			c2, err := c.CreateBucketIfNotExists([]byte("code"))
			if err != nil {
				return err
			}
			return c2.Put(codeHash[:], code)
		})
		if err != nil {
			panic(err)
		}
	}()
}

func SaveStorage(storageChange map[common.Hash]common.Hash, codeHash []byte, addrHash common.Hash, index model.BtIndex) {
	go func() {
		atomic.AddInt64(&parallelTask, 1)
		defer atomic.AddInt64(&parallelTask, -1)
		err := DataBases[Code].Batch(func(tx *bolt.Tx) error {
			b1, err := tx.CreateBucketIfNotExists(codeHash[:])
			if err != nil {
				return err
			}
			storageAddr := crypto.Keccak256Hash(codeHash, addrHash[:])
			b2, err := b1.CreateBucketIfNotExists(storageAddr[:])
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

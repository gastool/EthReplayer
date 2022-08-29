package database

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/research/model"
	"github.com/ethereum/go-ethereum/rlp"
	bolt "go.etcd.io/bbolt"
)

func GetStateAccount(bt model.BtIndex, addr common.Address) *types.StateAccount {
	var find bool
	var state model.AccountState
	err := DataBases[Account].View(func(tx *bolt.Tx) error {
		b := tx.Bucket(addr[:])
		if b != nil {
			c := b.Cursor()
			if k, v := c.Seek(bt.ToSearchKey(nil)); k != nil {
				find = true
				return rlp.DecodeBytes(v, &state)
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	if find && !state.Deleted {
		s := &types.StateAccount{
			Nonce:   state.Nonce,
			Balance: state.Balance,
		}
		if state.CodeHash != nil {
			s.Root = *state.Root
			s.CodeHash = *state.CodeHash
		}
		return s
	}
	return nil
}

func GetBlockInfo(bt model.BtIndex) *model.BlockInfo {
	key := bt.BlockToByte()
	var info *model.BlockInfo
	err := DataBases[Info].View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("block"))
		v := c.Get(key)
		if v != nil {
			info = &model.BlockInfo{}
			return rlp.DecodeBytes(v, info)
		}
		return nil
	})
	if err != nil {
		return nil
	}
	return info
}

func GetTxInfo(bt model.BtIndex) *model.TxInfo {
	key := bt.AllToByte()
	var info *model.TxInfo
	err := DataBases[Info].View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("tx"))
		v := c.Get(key)
		if v != nil {
			info = &model.TxInfo{}
			return rlp.DecodeBytes(v, info)
		}
		return nil
	})
	if err != nil {
		return nil
	}
	return info
}

func GetStorage(bt model.BtIndex, codeHash []byte, addrHash common.Hash) map[common.Hash]common.Hash {
	storage := make(map[common.Hash]common.Hash)
	//sk2 := bt.ToSortKey(nil)
	sk2 := bt.ToSearchKey(nil)
	err := DataBases[Code].View(func(tx *bolt.Tx) error {
		b := tx.Bucket(codeHash)
		if b != nil {
			storageAddr := crypto.Keccak256Hash(codeHash, addrHash[:])
			b2 := b.Bucket(storageAddr[:])
			if b2 != nil {
				c := b2.Cursor()
				for k, _ := c.First(); k != nil; k, _ = c.Next() {
					b3 := b2.Bucket(k)
					k2, v := b3.Cursor().Seek(sk2)
					if k2 == nil {
						continue
					}
					storage[common.BytesToHash(common.LeftPadBytes(k, common.HashLength))] =
						common.BytesToHash(common.LeftPadBytes(v, common.HashLength))
				}
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return storage
}

func GetStorageValue(bt model.BtIndex, codeHash []byte, addrHash common.Hash, key []byte) []byte {
	sk2 := bt.ToSearchKey(nil)
	key2 := common.TrimLeftZeroes(key)
	if len(key2) == 0 {
		key2 = []byte{0}
	}
	var value []byte
	DataBases[Code].View(func(tx *bolt.Tx) error {
		b := tx.Bucket(codeHash)
		if b != nil {
			storageAddr := crypto.Keccak256Hash(codeHash, addrHash[:])
			b2 := b.Bucket(storageAddr[:])
			if b2 != nil {
				b3 := b2.Bucket(key2)
				if b3 == nil {
					return nil
				}
				k2, v := b3.Cursor().Seek(sk2)
				if k2 != nil {
					value = v
					return nil
				}
			}
		}
		return nil
	})
	return value
}

func GetContractCode(codeHash []byte) ([]byte, error) {
	var bs []byte
	return bs, DataBases[Code].View(func(tx *bolt.Tx) error {
		b0 := tx.Bucket(codeHash)
		if b0 == nil {
			return nil
		}
		b := b0.Bucket([]byte("code"))
		if b == nil {
			return nil
		}
		bs = b.Get(codeHash)
		if len(bs) == 0 {
			return errors.New("code not found")
		}
		return nil
	})
}

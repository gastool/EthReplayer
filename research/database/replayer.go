package database

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/research/bbolt"
	"github.com/ethereum/go-ethereum/research/model"
	"github.com/ethereum/go-ethereum/rlp"
)

func GetReplayCount(bt model.BtIndex, addrHash common.Hash) uint32 {
	redeployIndex := uint32(0)
	if _, ok := Redeploy.Load(addrHash); ok {
		DataBases[CodeChange].View(func(tx *bbolt.Tx) error {
			c := tx.Bucket(addrHash[:]).Cursor()
			for k, v := c.First(); k != nil && bytes.Compare(k, bt.ToByte()) <= 0; k, v = c.Next() {
				var change model.CodeChange
				rlp.DecodeBytes(v, &change)
				if change.Redeploy {
					redeployIndex++
				}
			}
			return nil
		})
	}
	return redeployIndex
}

func GetStateAccount(bt model.BtIndex, addr common.Address) (s *types.StateAccount) {
	var find bool
	var state model.AccountState
	defer func() {
		if Debug {
			if s != nil {
				log.Info("debug account", "addr", addr, "nonce", state.Nonce, "balance", state.Balance,
					"codeHash", hex.EncodeToString(s.CodeHash), "deleted", state.Deleted)
			} else {
				log.Info("debug account", "addr", addr, "empty", "")
			}

		}
	}()
	err := DataBases[Account].View(func(tx *bbolt.Tx) error {
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
			s.Root = common.BytesToHash([]byte{0x1})
			s.CodeHash = *state.CodeHash
		}
		return s
	}
	return nil
}

func GetBlockInfo(bt model.BtIndex) *model.BlockInfo {
	key := bt.BlockToByte()
	var info *model.BlockInfo
	err := DataBases[Info].View(func(tx *bbolt.Tx) error {
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
	key := bt.ToByte()
	var info *model.TxInfo
	err := DataBases[Info].View(func(tx *bbolt.Tx) error {
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

func GetStorage(bt model.BtIndex, addrHash common.Hash, search bool) (s map[common.Hash]common.Hash) {
	storage := make(map[common.Hash]common.Hash)
	sk2 := bt.ToSearchKey(nil)
	if !search {
		sk2 = bt.ToSortKey(nil)
	}
	var storageAddr = addrHash
	redeployIndex := GetReplayCount(bt, addrHash)
	if redeployIndex > 0 {
		bs := make([]byte, 4)
		binary.LittleEndian.PutUint32(bs, redeployIndex)
		storageAddr = crypto.Keccak256Hash(addrHash[:], bs)
	}
	defer func() {
		if Debug {
			log.Info("debug storage", "addrHash", addrHash, "storage", s, "redeployIndex", redeployIndex)
		}
	}()
	err := DataBases[Storage].View(func(tx *bbolt.Tx) error {
		b2 := tx.Bucket(storageAddr[:])
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
		return nil
	})
	if err != nil {
		panic(err)
	}
	return storage
}

func GetStorageValue(bt model.BtIndex, addrHash common.Hash, key []byte) (v []byte) {
	sk2 := bt.ToSearchKey(nil)
	key2 := common.TrimLeftZeroes(key)
	if len(key2) == 0 {
		key2 = []byte{0}
	}
	var value []byte
	var storageAddr = addrHash
	redeployIndex := GetReplayCount(bt, addrHash)
	if redeployIndex > 0 {
		bs := make([]byte, 4)
		binary.LittleEndian.PutUint32(bs, redeployIndex)
		storageAddr = crypto.Keccak256Hash(addrHash[:], bs)
	}
	defer func() {
		if Debug {
			log.Info("debug storage", "key", hex.EncodeToString(key), "value", hex.EncodeToString(v),
				"addrHash", addrHash, "redeployIndex", redeployIndex)
		}
	}()
	DataBases[Storage].View(func(tx *bbolt.Tx) error {
		b2 := tx.Bucket(storageAddr[:])
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
		return nil
	})
	return value
}

func GetContractCode(codeHash []byte) ([]byte, error) {
	var bs []byte
	return bs, DataBases[Code].View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("code"))
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

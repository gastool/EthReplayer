package database

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/research/model"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/status-im/keycard-go/hexutils"
	bolt "go.etcd.io/bbolt"
	"testing"
)

func TestData(t *testing.T) {
	DataBases[Account].View(func(tx *bolt.Tx) error {
		c := tx.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			c2 := tx.Bucket(k).Cursor()
			t.Log(hexutils.BytesToHex(k))
			for k2, v2 := c2.First(); k2 != nil; k2, v2 = c2.Next() {
				var state model.AccountState
				err := rlp.DecodeBytes(v2, &state)
				if err != nil {
					panic(err)
				}
				fmt.Println(hexutils.BytesToHex(k2), state)
			}
		}
		return nil
	})
}

func TestGetData(t *testing.T) {
	s := GetStateAccount(model.BtIndex{
		BlockNumber: 560000,
		TxIndex:     100,
	}, common.HexToAddress("0x32be343b94f860124dc4fee278fdcbd38c102d88"))
	t.Log(s)
}

func TestGetStorage(t *testing.T) {
	bt := model.BtIndex{
		BlockNumber: 235555,
		TxIndex:     0,
	}
	addr := "0x1194e966965418c7d73a42cceeb254d875860356"
	s := GetStateAccount(bt, common.HexToAddress(addr))
	if len(s.CodeHash) == 0 {
		panic("")
	}
	storage := GetStorage(bt, s.CodeHash, common.HexToAddress(addr).Hash())
	for k, v := range storage {
		fmt.Println(hex.EncodeToString(k[:]), hex.EncodeToString(v[:]))
	}
}

func TestSortKey(t *testing.T) {
	bt := model.BtIndex{
		BlockNumber: 655,
		TxIndex:     1,
	}
	t.Log(hex.EncodeToString(bt.ToSortKey(nil)))
}

func TestSearchKey(t *testing.T) {
	bt := model.BtIndex{
		BlockNumber: 19235555,
		TxIndex:     1,
	}
	t.Log(hex.EncodeToString(bt.ToSearchKey(nil)))
}

func TestStorage(t *testing.T) {
	DataBases[Account].View(func(tx *bolt.Tx) error {
		c := tx.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			c2 := tx.Bucket(k).Cursor()
			t.Log(hexutils.BytesToHex(k))
			for k2, v2 := c2.First(); k2 != nil; k2, v2 = c2.Next() {
				var state model.AccountState
				err := rlp.DecodeBytes(v2, &state)
				if err != nil {
					panic(err)
				}
				fmt.Println(hexutils.BytesToHex(k2), state)
			}
		}
		return nil
	})
}

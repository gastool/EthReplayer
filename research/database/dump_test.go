package database

import (
	"bytes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/research/model"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/status-im/keycard-go/hexutils"
	bolt "go.etcd.io/bbolt"
	"testing"
)

var EmptyRoot = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")

func TestStorageVerify(t *testing.T) {
	var (
		diskdb = memorydb.New()
		triedb = trie.NewDatabase(diskdb)
	)

	DataBases[Account].View(func(tx *bolt.Tx) error {
		c := tx.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			c2 := tx.Bucket(k).Cursor()
			for k2, v2 := c2.First(); k2 != nil; k2, v2 = c2.Next() {
				var state model.AccountState
				err := rlp.DecodeBytes(v2, &state)
				if err != nil {
					panic(err)
				}
				bt := model.KeyToBtIndex(k2)
				if state.CodeHash == nil {
					continue
				}
				storageTrie, err := trie.NewSecure(common.Hash{}, EmptyRoot, triedb)
				sm := GetStorage(bt, *state.CodeHash, crypto.Keccak256Hash(k))
				for k, v := range sm {
					if (v == common.Hash{}) {
						err := storageTrie.TryDelete(k[:])
						if err != nil {
							panic(err)
						}
						continue
					}
					var value []byte
					value, _ = rlp.EncodeToBytes(common.TrimLeftZeroes(v[:]))
					err = storageTrie.TryUpdate(k[:], value)
					if err != nil {
						panic(err)
					}
				}
				root := storageTrie.Hash()
				if !bytes.Equal((*state.Root)[:], root[:]) {
					t.Log("no pass", hexutils.BytesToHex(k), bt)

					//panic(hex.EncodeToString(k))
				} else {
					//t.Log(hexutils.BytesToHex(k))
				}
			}
		}
		return nil
	})
}

func TestAccountVerify(t *testing.T) {
	var (
		diskdb = memorydb.New()
		triedb = trie.NewDatabase(diskdb)
	)

	DataBases[Account].View(func(tx *bolt.Tx) error {
		c := tx.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			c2 := tx.Bucket(k).Cursor()
			for k2, v2 := c2.First(); k2 != nil; k2, v2 = c2.Next() {
				var state model.AccountState
				err := rlp.DecodeBytes(v2, &state)
				if err != nil {
					panic(err)
				}
				bt := model.KeyToBtIndex(k2)
				if state.CodeHash == nil {
					continue
				}
				storageTrie, err := trie.NewSecure(common.Hash{}, EmptyRoot, triedb)
				sm := GetStorage(bt, *state.CodeHash, crypto.Keccak256Hash(k))
				for k, v := range sm {
					if (v == common.Hash{}) {
						err := storageTrie.TryDelete(k[:])
						if err != nil {
							panic(err)
						}
						continue
					}
					var value []byte
					value, _ = rlp.EncodeToBytes(common.TrimLeftZeroes(v[:]))
					err = storageTrie.TryUpdate(k[:], value)
					if err != nil {
						panic(err)
					}
				}
				root := storageTrie.Hash()
				if !bytes.Equal((*state.Root)[:], root[:]) {
					t.Log("no pass", hexutils.BytesToHex(k), bt)

					//panic(hex.EncodeToString(k))
				} else {
					//t.Log(hexutils.BytesToHex(k))
				}
			}
		}
		return nil
	})
}

func TestStorageVerifySingle(t *testing.T) {
	var (
		diskdb = memorydb.New()
		triedb = trie.NewDatabase(diskdb)
	)
	DataBases[Account].View(func(tx *bolt.Tx) error {
		addr := common.HexToAddress("FBC128067A2FE13C11BBC6CC55F29AA1F27630DA")
		c2 := tx.Bucket(addr[:]).Cursor()
		for k2, v2 := c2.First(); k2 != nil; k2, v2 = c2.Next() {
			var state model.AccountState
			err := rlp.DecodeBytes(v2, &state)
			if err != nil {
				panic(err)
			}
			bt := model.KeyToBtIndex(k2)
			if state.CodeHash == nil {
				continue
			}
			if bt.BlockNumber != 62737 || bt.TxIndex != 0 {
				continue
			}
			storageTrie, err := trie.NewSecure(common.Hash{}, EmptyRoot, triedb)
			t.Log(bt)
			sm := GetStorage(bt, *state.CodeHash, crypto.Keccak256Hash(addr[:]))
			for k, v := range sm {
				t.Log(k.String(), v.String())
				if (v == common.Hash{}) {
					err := storageTrie.TryDelete(k[:])
					if err != nil {
						panic(err)
					}
					continue
				}
				var value []byte
				value, _ = rlp.EncodeToBytes(common.TrimLeftZeroes(v[:]))
				err = storageTrie.TryUpdate(k[:], value)
				if err != nil {
					panic(err)
				}
			}
			root := storageTrie.Hash()
			if !bytes.Equal((*state.Root)[:], root[:]) {
				t.Log("no pass", bt)
				//panic(hex.EncodeToString(k))
			} else {
				t.Log("pass", bt)
			}
		}
		return nil
	})

}

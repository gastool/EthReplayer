package replay

import (
	"github.com/ethereum/go-ethereum/research/database"
	"github.com/ethereum/go-ethereum/research/model"
	bolt "go.etcd.io/bbolt"
	"testing"
)

func TestReplay(t *testing.T) {
	database.DataBases[database.Account].View(func(tx *bolt.Tx) error {
		c := tx.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			c2 := tx.Bucket(k).Cursor()
			for k2, _ := c2.First(); k2 != nil; k2, _ = c2.Next() {
				bt := model.KeyToBtIndex(k2)
				if bt.TxIndex == 65535 {
					continue
				}
				Replay(uint64(bt.BlockNumber), int(bt.TxIndex))
			}
		}
		return nil
	})
}

func TestReplay2(t *testing.T) {
	Replay(56187, 0)
}

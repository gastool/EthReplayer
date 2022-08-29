package replay

import (
	"github.com/ethereum/go-ethereum/research/database"
	"github.com/ethereum/go-ethereum/research/model"
	bolt "go.etcd.io/bbolt"
	"log"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestReplay(t *testing.T) {
	start := time.Now().Unix()
	count := int64(0)
	taskNumber := int64(500)
	wg := &sync.WaitGroup{}
	sum := 0
	database.DataBases[database.Account].View(func(tx *bolt.Tx) error {
		c := tx.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			if sum >= 100 {
				return nil
			}
			c2 := tx.Bucket(k).Cursor()
			for k2, _ := c2.First(); k2 != nil; k2, _ = c2.Next() {
				bt := model.KeyToBtIndex(k2)
				if bt.TxIndex == 65535 {
					continue
				}
				wg.Add(1)
				for atomic.LoadInt64(&count) >= taskNumber {
					time.Sleep(200 * time.Millisecond)
				}
				atomic.AddInt64(&count, 1)
				go func() {
					Replay(uint64(bt.BlockNumber), int(bt.TxIndex))
					wg.Done()
					atomic.AddInt64(&count, -1)
				}()
			}
			sum++

		}
		return nil
	})
	wg.Wait()
	end := time.Now().Unix()
	log.Println(end - start)
}

func TestReplay2(t *testing.T) {
	Replay(696021, 1)
}

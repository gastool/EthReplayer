package database

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/log"
	bolt "go.etcd.io/bbolt"
	"io/ioutil"
	"os"
	"sync/atomic"
	"time"
)

type DataType int

var DataBases map[DataType]*bolt.DB

var configText = `
{
  "dir": "F:/",
  "Names": [
    "account",
    "code",
    "info"
  ]
}`

type DBConfig struct {
	Dir   string   `json:"dir"`
	Names []string `json:"names"`
}

func init() {
	bs, err := ioutil.ReadFile("config.json")
	if err != nil {
		//panic(err)
		bs = []byte(configText)
	}
	var c DBConfig
	err = json.Unmarshal(bs, &c)
	if err != nil {
		panic(err)
	}
	DataBases = make(map[DataType]*bolt.DB)
	for i := Account; i <= Info; i++ {
		db, err := bolt.Open(c.Dir+c.Names[i]+".db", os.ModePerm, nil)
		if err != nil {
			panic(err)
		}
		db.MaxBatchSize = 5000
		db.MaxBatchDelay = 500 * time.Millisecond
		DataBases[i] = db
	}
}

func Close() {
	for {
		tn := atomic.LoadInt64(&saveTask)
		if tn > 0 {
			log.Info("database closing", "saveTask", tn)
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	for _, v := range DataBases {
		v.Close()
	}
	log.Info("database closed")
}

const (
	Account DataType = iota
	Code
	Info
)

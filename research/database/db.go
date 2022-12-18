package database

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/research/bbolt"
	"io/ioutil"
	"os"
	"sync/atomic"
	"time"
)

type DataType int

var DataBases map[DataType]*bbolt.DB

var configText = `
{
  "dir": "F:/",
  "Names": [
    "account",
    "code",
    "codeChange",
    "storage",
    "info"
  ]
}`

type DBConfig struct {
	Dir   string   `json:"dir"`
	Names []string `json:"names"`
}

var ReplayMode bool

var Debug bool
var MaxBatchDelay = 60 * time.Second
var MaxBatchSize = 1000000

func Init() {
	var e error
	if e != nil {
		panic(e)
	}
	bs, err := ioutil.ReadFile("config.json")
	if err != nil {
		bs = []byte(configText)
	}
	var c DBConfig
	err = json.Unmarshal(bs, &c)
	if err != nil {
		panic(err)
	}
	if ReplayMode {
		log.Info("replay db", "readonly", true)
	}
	DataBases = make(map[DataType]*bbolt.DB)
	for i := Account; i <= Info; i++ {
		var option *bbolt.Options
		if ReplayMode {
			option = &bbolt.Options{ReadOnly: true}
		}
		db, err := bbolt.Open(c.Dir+c.Names[i]+".db", os.ModePerm, option)
		if err != nil {
			panic(err)
		}
		db.MaxBatchSize = MaxBatchSize
		db.MaxBatchDelay = MaxBatchDelay
		DataBases[i] = db
		if i == Info && !ReplayMode {
			err = db.Update(func(tx *bbolt.Tx) error {
				_, err := tx.CreateBucketIfNotExists([]byte("tx"))
				if err != nil {
					return err
				}
				_, err = tx.CreateBucketIfNotExists([]byte("block"))
				if err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				panic(err)
			}
		}
		if i == Code && !ReplayMode {
			err = db.Update(func(tx *bbolt.Tx) error {
				_, err := tx.CreateBucketIfNotExists([]byte("code"))
				if err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				panic(err)
			}
		}
	}
	InitRedeploy()
}

func Close() {
	for {
		tn := atomic.LoadInt64(&saveTask)
		if tn > 0 {
			log.Info("database closing", "saveTask=", tn)
			time.Sleep(30 * time.Second)
		} else {
			break
		}
	}
	time.Sleep(MaxBatchDelay) // Wait batch commit
	for _, v := range DataBases {
		v.Close()
	}
	log.Info("database closed")
}

const (
	Account DataType = iota
	Code
	CodeChange
	Storage
	Info
)

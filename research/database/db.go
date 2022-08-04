package database

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/log"
	bolt "go.etcd.io/bbolt"
	"io/ioutil"
	"os"
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
		DataBases[i] = db
	}
}

func Close() {
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

package replay

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/research/bbolt"
	"github.com/ethereum/go-ethereum/research/database"
	"github.com/ethereum/go-ethereum/research/model"
	"github.com/gammazero/workerpool"
	"gopkg.in/urfave/cli.v1"
	"log"
	"strconv"
	"sync"
	"time"
)

var rangeReplay bool
var debug bool

var workers int

var rangeFlag = &cli.BoolFlag{Name: "range", Destination: &rangeReplay}

var debugFlag = &cli.BoolFlag{Name: "debug", Destination: &debug}
var workerFlag = &cli.IntFlag{Name: "workers", Destination: &workers}

var ReplayCommand = cli.Command{
	Action:      replayCmd,
	Name:        "replay",
	Flags:       []cli.Flag{rangeFlag, workerFlag, debugFlag},
	Usage:       "replay blockNumber txIndex",
	ArgsUsage:   "<blockNumber> <txIndex>",
	Description: `replay tx`,
}

func replayCmd(ctx *cli.Context) error {
	database.ReplayMode = true
	database.Init()
	if debug {
		database.Debug = true
	}
	if len(ctx.Args()) == 0 {
		replayAllTx()
		return nil
	}
	p1, _ := strconv.Atoi(ctx.Args()[0])
	p2, _ := strconv.Atoi(ctx.Args()[1])
	if rangeReplay {
		if len(ctx.Args())%2 != 0 {
			return errors.New("invalid block range")
		}
		for i := 0; i < len(ctx.Args())-1; i = i + 2 {
			p1, _ = strconv.Atoi(ctx.Args()[i])
			p2, _ = strconv.Atoi(ctx.Args()[i+1])
			replayRange(p1, p2)
		}
		return nil
	} else {
		return Replay(uint64(p1), p2)
	}
}

func replayAllTx() {
	start := time.Now()
	sum := 0
	lb := uint32(0)
	lock := &sync.Mutex{}
	if workers <= 1 {
		workers = 1
	}
	wp := workerpool.New(workers)

	database.DataBases[database.Info].View(func(tx *bbolt.Tx) error {
		c := tx.Bucket([]byte("tx")).Cursor()
		last := time.Now()
		txSum := 0
		blockSum := uint32(0)
		lastBlock := uint32(0)

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			k2 := make([]byte, len(k))
			copy(k2, k)
			wp.Submit(func() {
				bt := model.KeyToBtIndex(k2)
				if bt.BlockNumber == 0 && bt.TxIndex == 0 {
					return
				}
				err := Replay(uint64(bt.BlockNumber), int(bt.TxIndex))
				lock.Lock()
				if lastBlock != bt.BlockNumber {
					blockSum++
					lastBlock = bt.BlockNumber
				}
				sum++
				txSum++
				lb2 := bt.BlockNumber / 10000
				if lb2 != lb {
					elapsed := time.Since(last).Seconds()
					lb = lb2
					last = time.Now()
					fmt.Printf("elapsed time: %v, number = %v\n", time.Since(start).Round(time.Millisecond), bt.BlockNumber)
					fmt.Printf("%.2f blk/s   %.2f tx/s\n", float64(blockSum)/elapsed, float64(txSum)/elapsed)
					txSum = 0
					blockSum = 0
				}
				lock.Unlock()
				if err != nil {
					log.Println(err, bt)
				}

			})
		}
		return nil
	})
	wp.StopWait()
	fmt.Println("cost time:", time.Since(start).Round(time.Millisecond))
	fmt.Printf("all tx:%d", sum)
}

func replayRange(bs int, be int) error {
	lock := &sync.Mutex{}
	if workers <= 1 {
		workers = 1
	}
	wp := workerpool.New(workers)
	fmt.Printf("replay range start:%d end:%d\n", bs, be)
	start := time.Now()
	sum := 0
	sbt := model.BtIndex{BlockNumber: uint32(bs)}
	startByte := sbt.ToByte()
	ebt := model.BtIndex{BlockNumber: uint32(be + 1)}
	endByte := ebt.ToByte()
	lb := uint32(0)

	err := database.DataBases[database.Info].View(func(tx *bbolt.Tx) error {
		batch := uint32(10000)
		c := tx.Bucket([]byte("tx")).Cursor()
		last := time.Now()
		txSum := 0
		blockSum := uint32(0)
		lastBlock := uint32(0)
		for k, _ := c.Seek(startByte); k != nil && bytes.Compare(endByte, k) > 0; k, _ = c.Next() {
			k2 := make([]byte, len(k))
			copy(k2, k)
			wp.Submit(func() {
				bt := model.KeyToBtIndex(k2)
				if bt.BlockNumber == 0 && bt.TxIndex == 0 {
					return
				}
				err := Replay(uint64(bt.BlockNumber), int(bt.TxIndex))
				lock.Lock()
				lb2 := bt.BlockNumber / batch
				if lb == 0 {
					lb = lb2
				}
				if lastBlock != bt.BlockNumber {
					blockSum++
					lastBlock = bt.BlockNumber
				}
				sum++
				txSum++
				if lb2 > lb {
					elapsed := time.Since(last).Seconds()
					last = time.Now()
					d := time.Since(start)
					fmt.Printf("elapsed time: %v, blk  = %v, tx =%v,elapsed millisecond=%v  \n", d.Round(time.Millisecond), bt.BlockNumber, txSum, d.Milliseconds())
					fmt.Printf("%.2f blk/s   %.2f tx/s\n", float64(blockSum)/elapsed, float64(txSum)/elapsed)
					txSum = 0
					blockSum = 0
					lb = lb2
				}
				lock.Unlock()
				if err != nil {
					log.Println("error happened:", err, bt)
					return
				}
			})
		}
		return nil
	})
	wp.StopWait()
	d := time.Since(start)
	fmt.Println("cost time:", d.Round(time.Millisecond), d.Milliseconds())
	fmt.Printf("\nall tx:%d", sum)
	return err
}

package replay

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/research/bbolt"
	"github.com/ethereum/go-ethereum/research/database"
	"github.com/ethereum/go-ethereum/research/model"
	"log"
	"math/big"
)

func NewReplayStateDB() *state.StateDB {
	sdb := state.NewDatabase(rawdb.NewMemoryDatabase())
	statedb, _ := state.New(common.Hash{}, sdb, nil)
	//rb := state.NewReplayDB(sdb.TrieDB())
	rb := state.NewReplayLazyDB(sdb.TrieDB())
	statedb.ReplayDb = rb
	return statedb
}

func ReplayAddr(address common.Address) {
	database.DataBases[database.Account].View(func(tx *bbolt.Tx) error {
		c := tx.Bucket(address.Bytes())
		if c == nil {
			return errors.New("no such address")
		}
		cur := c.Cursor()
		for k, _ := cur.First(); k != nil; k, _ = cur.Next() {
			bt := model.SortKeyToBtIndex(k)
			Replay(uint64(bt.BlockNumber), int(bt.TxIndex))
		}
		return nil
	})
}

func Replay(block uint64, txIndex int) error {
	var (
		chainConfig *params.ChainConfig
		statedb     = NewReplayStateDB()
		gaspool     = new(core.GasPool)
	)
	bt := model.BtIndex{
		BlockNumber: uint32(block),
		TxIndex:     uint16(txIndex),
	}
	vmConfig := vm.Config{
		Debug: false,
	}
	chainConfig = &params.ChainConfig{}
	chainConfig = params.MainnetChainConfig
	chainConfig.DAOForkSupport = false
	database.ReplayMode = true
	getHash := func(num uint64) common.Hash {
		b := database.GetBlockInfo(model.BtIndex{BlockNumber: uint32(num)})
		if b == nil {
			return common.Hash{}
		}
		return b.BlockHash
	}

	be := database.GetBlockInfo(bt)
	txInfo := database.GetTxInfo(bt)
	if be == nil || txInfo == nil {
		//log.Printf("invalid blockNumber:%d or tx:%d", block, txIndex)
		return fmt.Errorf("invalid blockNumber:%d or tx:%d", block, txIndex)
	}
	gaspool.AddGas(be.GasLimit)
	vmContext := vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		Coinbase:    be.Coinbase,
		BlockNumber: new(big.Int).SetUint64(block),
		Time:        new(big.Int).SetUint64(be.Time),
		Difficulty:  be.Difficulty,
		GasLimit:    be.GasLimit,
		GetHash:     getHash,
	}
	msg := txInfo.AsMessage()
	statedb.Prepare(txInfo.Hash, txIndex)
	statedb.SetBtIndex(block, txIndex)
	txContext := vm.TxContext{
		Origin:   msg.From(),
		GasPrice: msg.GasPrice(),
	}
	evm := vm.NewEVM(vmContext, txContext, statedb, chainConfig, vmConfig)
	result, err := core.ApplyMessage(evm, msg, gaspool)
	if err != nil {
		log.Println("[error]", block, txIndex, err)
	}
	status := uint64(0)
	if result.Failed() {
		status = types.ReceiptStatusFailed
	} else {
		status = types.ReceiptStatusSuccessful
	}
	var contractAddr common.Address
	if msg.To() == nil {
		contractAddr = crypto.CreateAddress(evm.TxContext.Origin, txInfo.Nonce)
	}
	logs := statedb.GetLogs(txInfo.Hash, be.BlockHash)
	hash := model.GenerateExecuteHash(logs, result.UsedGas, status, contractAddr)
	if !bytes.Equal(hash[:], txInfo.ResultHash[:]) {
		logStr := ""
		for _, v := range logs {
			topics := "["
			for _, v2 := range v.Topics {
				topics += v2.String() + " "
			}
			topics += "]"
			logStr += fmt.Sprintf("topics:%s data:%s", topics, hex.EncodeToString(v.Data))
		}
		fmt.Printf("[error info] gasUsed:%d err:%s  returnData:%s\nlog:%s\n", result.UsedGas,
			result.Err, hex.EncodeToString(result.ReturnData), logStr)
		return fmt.Errorf("error at %d %d %s", block, txIndex, txInfo.Hash.String())
	}
	return nil
}

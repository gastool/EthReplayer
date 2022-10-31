package model

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

type AccountState struct {
	Nonce    uint64
	Balance  *big.Int
	CodeHash *[]byte `rlp:"nil"`
	Deleted  bool
}

type TxInfo struct {
	To         *common.Address `rlp:"nil"`
	From       common.Address
	Nonce      uint64
	Amount     *big.Int `rlp:"nil"`
	GasLimit   uint64
	GasPrice   *big.Int `rlp:"nil"`
	GasTipCap  *big.Int `rlp:"nil"`
	GasFeeCap  *big.Int `rlp:"nil"`
	Data       []byte
	Hash       common.Hash
	ResultHash common.Hash
}

type BlockInfo struct {
	Coinbase   common.Address
	GasLimit   uint64
	Difficulty *big.Int
	Number     *big.Int
	Time       uint64
	BlockHash  common.Hash
}

type CodeChange struct {
	Delete   bool
	Redeploy bool
}

func (t *TxInfo) AsMessage() types.Message {
	return types.NewMessage(t.From, t.To, t.Nonce, t.Amount, t.GasLimit, t.GasPrice, t.GasFeeCap, t.GasTipCap, t.Data, nil, false)
}

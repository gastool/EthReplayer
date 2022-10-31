package model

import (
	"encoding/binary"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
)

func GenerateExecuteHash(logs []*types.Log, usedGas uint64, status uint64, contractAddress common.Address) common.Hash {
	bs, err := rlp.EncodeToBytes(&logs)
	s1 := make([]byte, 8)
	binary.LittleEndian.PutUint64(s1, usedGas)
	bs = append(bs, s1...)
	s2 := make([]byte, 8)
	binary.LittleEndian.PutUint64(s2, status)
	bs = append(bs, s2...)
	bs = append(bs, contractAddress.Bytes()...)
	if err != nil {
		panic(err)
	}
	return crypto.Keccak256Hash(bs)
}

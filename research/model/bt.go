package model

import "encoding/binary"

type BtIndex struct {
	BlockNumber uint32
	TxIndex     uint16
}

func (bt *BtIndex) ToByte() []byte {
	b := make([]byte, 6)
	v := bt.BlockNumber
	b[0] = byte(v >> 24)
	b[1] = byte(v >> 16)
	b[2] = byte(v >> 8)
	b[3] = byte(v)
	v2 := bt.TxIndex
	b[4] = byte(v2 >> 8)
	b[5] = byte(v2)
	return b
}

func (bt *BtIndex) BlockToByte() []byte {
	b := make([]byte, 4)
	v := bt.BlockNumber
	b[0] = byte(v >> 24)
	b[1] = byte(v >> 16)
	b[2] = byte(v >> 8)
	b[3] = byte(v)
	return b
}

func (bt *BtIndex) ToSortKey(in []byte) []byte {
	b := make([]byte, 6)
	v := ^bt.BlockNumber
	b[0] = byte(v >> 24)
	b[1] = byte(v >> 16)
	b[2] = byte(v >> 8)
	b[3] = byte(v)
	v2 := ^bt.TxIndex
	b[4] = byte(v2 >> 8)
	b[5] = byte(v2)
	return append(b, in...)
}

func (bt *BtIndex) ToSearchKey(in []byte) []byte {
	v := ^bt.BlockNumber
	v2 := 1 + ^bt.TxIndex
	if bt.TxIndex == 0 {
		v = v + 1
		v2 = 0
	}
	b := make([]byte, 6)
	b[0] = byte(v >> 24)
	b[1] = byte(v >> 16)
	b[2] = byte(v >> 8)
	b[3] = byte(v)
	b[4] = byte(v2 >> 8)
	b[5] = byte(v2)
	return append(b, in...)
}

func SortKeyToBtIndex(key []byte) BtIndex {
	var bt BtIndex
	if len(key) != 6 {
		return bt
	}
	b1 := make([]byte, 4)
	for i := 0; i < 4; i++ {
		b1[i] = key[i]
	}
	bt.BlockNumber = ^binary.BigEndian.Uint32(b1)
	b2 := make([]byte, 2)
	for i := 4; i < 6; i++ {
		b2[i-4] = key[i]
	}
	bt.TxIndex = ^binary.BigEndian.Uint16(b2)
	return bt
}

func KeyToBtIndex(key []byte) BtIndex {
	var bt BtIndex
	if len(key) != 6 {
		return bt
	}
	b1 := make([]byte, 4)
	for i := 0; i < 4; i++ {
		b1[i] = key[i]
	}
	bt.BlockNumber = binary.BigEndian.Uint32(b1)
	b2 := make([]byte, 2)
	for i := 4; i < 6; i++ {
		b2[i-4] = key[i]
	}
	bt.TxIndex = binary.BigEndian.Uint16(b2)
	return bt
}

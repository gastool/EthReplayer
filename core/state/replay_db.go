package state

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/research/database"
	"github.com/ethereum/go-ethereum/research/model"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	lru "github.com/hashicorp/golang-lru"
)

type ContextDatabase interface {
	Database
	SetBtIndex(bt model.BtIndex)
	SetCurrentCodeHash(codeHash *[]byte)
	IsLazy() bool
}

type CacheTrie struct {
	bt   model.BtIndex
	trie Trie
}

type LazyTrie struct {
	bt       model.BtIndex
	addrHash common.Hash
}

func (l *LazyTrie) SetContext(bt model.BtIndex, addrHash common.Hash) {
	l.bt = bt
	l.addrHash = addrHash
}

func (l *LazyTrie) GetKey([]byte) []byte {
	return nil
}

func (l *LazyTrie) TryGet(key []byte) ([]byte, error) {
	value := database.GetStorageValue(l.bt, l.addrHash, key)
	return value, nil
}

func (l *LazyTrie) TryUpdateAccount(key []byte, account *types.StateAccount) error {
	return nil

}
func (l *LazyTrie) TryUpdate(key, value []byte) error {
	return nil

}
func (l *LazyTrie) TryDelete(key []byte) error {
	return nil

}
func (l *LazyTrie) Hash() common.Hash {
	return common.Hash{}

}
func (l *LazyTrie) Commit(onleaf trie.LeafCallback) (common.Hash, int, error) {
	return common.Hash{}, 0, nil

}
func (l *LazyTrie) NodeIterator(startKey []byte) trie.NodeIterator {
	return nil

}
func (l *LazyTrie) Prove(key []byte, fromLevel uint, proofDb ethdb.KeyValueWriter) error {
	return nil
}

type RepDB struct {
	db              *trie.Database
	currentCodeHash *[]byte
	bt              model.BtIndex
	cacheTrie       *lru.Cache
	lazyStorageDB   bool
}

func NewReplayDB(db *trie.Database) *RepDB {
	cache, e := lru.New(100000)
	if e != nil {
		panic(e)
	}
	return &RepDB{
		db:        db,
		cacheTrie: cache,
	}
}

func NewReplayLazyDB(db *trie.Database) *RepDB {
	return &RepDB{
		db:            db,
		lazyStorageDB: true,
	}
}

func (r *RepDB) OpenTrie(root common.Hash) (Trie, error) {
	tr, err := trie.NewSecure(common.Hash{}, root, r.db)
	if err != nil {
		return nil, err
	}
	return tr, nil
}
func (r *RepDB) IsLazy() bool {
	return r.lazyStorageDB
}

func (r *RepDB) OpenStorageTrie(addrHash, root common.Hash) (Trie, error) {
	if r.lazyStorageDB {
		return &LazyTrie{
			addrHash: addrHash,
		}, nil
	}
	if value, ok := r.cacheTrie.Get(addrHash); ok {
		v := value.(CacheTrie)
		if v.bt.BlockNumber == r.bt.BlockNumber && v.bt.TxIndex == r.bt.TxIndex {
			return v.trie, nil
		}
	}
	tr, _ := trie.NewSecure(common.Hash{}, common.Hash{}, r.db)
	storageMap := database.GetStorage(r.bt, addrHash, true)
	for k, v := range storageMap {
		if v == (common.Hash{}) {
			tr.TryDelete(k[:])
		} else {
			value, _ := rlp.EncodeToBytes(common.TrimLeftZeroes(v[:]))
			tr.TryUpdate(k[:], value[:])
		}
	}
	if tr.Hash() != root {
		panic("invalid storage")
	}
	r.cacheTrie.Add(addrHash, &CacheTrie{
		bt:   r.bt,
		trie: tr,
	})
	return tr, nil
}

func (r *RepDB) CopyTrie(t Trie) Trie {
	switch t := t.(type) {
	case *trie.SecureTrie:
		return t.Copy()
	default:
		panic(fmt.Errorf("unknown trie type %T", t))
	}
}
func (r *RepDB) ContractCode(addrHash, codeHash common.Hash) ([]byte, error) {
	return database.GetContractCode(codeHash[:])
}

func (r *RepDB) ContractCodeSize(addrHash, codeHash common.Hash) (int, error) {
	bs, err := r.ContractCode(addrHash, codeHash)
	if err != nil {
		return 0, err
	}
	return len(bs), nil
}

func (r *RepDB) TrieDB() *trie.Database {
	return r.db
}

func (r *RepDB) SetBtIndex(bt model.BtIndex) {
	r.bt = bt
}

func (r *RepDB) GetStateObject(addr common.Address, bt model.BtIndex) *types.StateAccount {
	return database.GetStateAccount(bt, addr)
}

func (r *RepDB) SetCurrentCodeHash(hash *[]byte) {
	r.currentCodeHash = hash
}

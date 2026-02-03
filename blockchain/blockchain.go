package blockchain

import (
	"fmt"

	"github.com/dgraph-io/badger/v4"
)

type BlockchainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

const (
	dbPath = "./tmp/blocks"
)

type Blockchain struct {
	LastHash []byte
	Database *badger.DB
}

func InitBlockchain() *Blockchain {
	var lastHash []byte
	opts := badger.DefaultOptions(dbPath)
	opts.Dir = dbPath
	opts.ValueDir = dbPath
	db, err := badger.Open(opts)
	if err != nil {
		Handle(err)
	}
	err = db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			//execute genesis
			fmt.Println("not blockchain found - Creating genesis")
			genesis := Genesis()
			err = txn.Set(genesis.Hash, genesis.Serialize())
			Handle(err)
			err = txn.Set([]byte("lh"), genesis.Hash)
			Handle(err)
			lastHash = genesis.Hash
			return err
		} else {
			item, err := txn.Get([]byte("lh"))
			Handle(err)
			err1 := item.Value(func(val []byte) error {
				lastHash = append([]byte{}, val...)
				return nil
			})
			Handle(err1)
			return err
		}
	})
	Handle(err)
	blockchain := Blockchain{
		LastHash: lastHash,
		Database: db,
	}
	return &blockchain
}
func (chain *Blockchain) AddBlock(data string) {
	var lastHash []byte
	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		err1 := item.Value(func(val []byte) error {
			lastHash = append([]byte{}, val...)
			return nil
		})
		Handle(err1)
		return err
	})
	Handle(err)
	NewBlock := CreateBlock(data, lastHash)
	err = chain.Database.Update(func(txn *badger.Txn) error {
		err = txn.Set(NewBlock.Hash, NewBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), NewBlock.Hash)
		chain.LastHash = NewBlock.Hash
		return err
	})
	Handle(err)
}

func (chain *Blockchain) Iterator() *BlockchainIterator {
	iter := &BlockchainIterator{
		CurrentHash: chain.LastHash,
		Database:    chain.Database,
	}
	return iter
}

func (iter *BlockchainIterator) Next() *Block {
	var block *Block
	var blockData []byte
	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		Handle(err)
		err1 := item.Value(func(val []byte) error {
			blockData = append([]byte{}, val...)
			return nil
		})
		Handle(err1)
		block = Deserialize(blockData)
		return err
	})
	Handle(err)
	iter.CurrentHash = block.PrevHash
	return block
}

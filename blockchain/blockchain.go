package blockchain

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/dgraph-io/badger/v4"
)

type BlockchainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

const (
	dbPath      = "./tmp/blocks"
	dbFile      = "./tmp/blocks/MANIFEST"
	genesisData = "First transaction from genesis data"
)

type Blockchain struct {
	LastHash []byte
	Database *badger.DB
}

func DbExist() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

func ContinueBlockchain(address string) *Blockchain {
	if DbExist() == false {
		fmt.Println("Blockchain not exist, please create one")
		runtime.Goexit()
	}
	var lastHash []byte
	opts := badger.DefaultOptions(dbPath)
	opts.Dir = dbPath
	opts.ValueDir = dbPath
	db, err := badger.Open(opts)
	if err != nil {
		Handle(err)
	}
	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		err1 := item.Value(func(val []byte) error {
			lastHash = append([]byte{}, val...)
			return nil
		})
		if err1 != nil {
			log.Panic(err)
		}
		return err
	})
	chain := Blockchain{
		LastHash: lastHash,
		Database: db,
	}
	return &chain
}

func (chain *Blockchain) FindUnspentTransactions(address string) []Transaction {
	var unspentTransactions []Transaction
	spentTXOs := make(map[string][]int) // {txID: [indices de outputs gastos]}

	iter := chain.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

			// 1️⃣ Marca outputs gastos por meus inputs
			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					if in.CanUnlock(address) {
						// ⚡ use o ID da transação de onde veio o output
						inputTxID := hex.EncodeToString(in.ID)
						spentTXOs[inputTxID] = append(spentTXOs[inputTxID], in.Out)
					}
				}
			}

		Outputs:
			for outIdx, out := range tx.Outputs {
				// se esse output já foi gasto, ignora
				if spentOutputs, ok := spentTXOs[txID]; ok {
					for _, spentOut := range spentOutputs {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				// se é meu e não foi gasto → unspent
				if out.CanBeUnlock(address) {
					unspentTransactions = append(unspentTransactions, *tx)
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return unspentTransactions
}

func (b *Blockchain) FindUnspentTransactionOutput(address string) []TxOutputs {
	var unspentTransactionOutput []TxOutputs
	unspentTransaction := b.FindUnspentTransactions(address)
	for _, item := range unspentTransaction {
		for _, out := range item.Outputs {
			if out.CanBeUnlock(address) {
				unspentTransactionOutput = append(unspentTransactionOutput, out)
			}
		}
	}
	return unspentTransactionOutput
}

func (b *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTransactions := b.FindUnspentTransactions(address)
	accumulated := 0
Work:
	for _, item := range unspentTransactions {
		transactionId := hex.EncodeToString(item.ID)
		for outIndex, out := range item.Outputs {
			if out.CanBeUnlock(address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[transactionId] = append(unspentOutputs[transactionId], outIndex)
				if accumulated >= amount {
					break Work
				}
			}
		}
	}
	return accumulated, unspentOutputs
}

func InitBlockchain(address string) *Blockchain {
	var lastHash []byte
	if DbExist() {
		fmt.Println("Blockchain already exist")
		runtime.Goexit()
	}
	opts := badger.DefaultOptions(dbPath)
	opts.Dir = dbPath
	opts.ValueDir = dbPath
	db, err := badger.Open(opts)
	if err != nil {
		Handle(err)
	}
	err = db.Update(func(txn *badger.Txn) error {
		cbtx := CoinBaseTx(address, genesisData)
		Genesis := Genesis(cbtx)
		fmt.Println("Genesis have created")
		err := txn.Set(Genesis.Hash, Genesis.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), Genesis.Hash)
		Handle(err)
		lastHash = Genesis.Hash
		return err
	})
	Handle(err)
	blockchain := Blockchain{
		LastHash: lastHash,
		Database: db,
	}
	return &blockchain
}
func (chain *Blockchain) AddBlock(transaction []*Transaction) {
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
	NewBlock := CreateBlock(transaction, lastHash)
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

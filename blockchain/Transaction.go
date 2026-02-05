package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

type Transaction struct {
	ID      []byte
	Inputs  []TxInputs
	Outputs []TxOutputs
}

func (tx *Transaction) SetId() {
	var encoded bytes.Buffer
	var hash [32]byte
	encode := gob.NewEncoder(&encoded)
	if err := encode.Encode(tx); err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

type TxOutputs struct {
	Value  int
	PubKey string
}

type TxInputs struct {
	ID  []byte
	Out int
	Sig string
}

func CoinBaseTx(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Tokens to:%s", to)
	}
	txInput := TxInputs{
		ID:  []byte{},
		Out: -1,
		Sig: data,
	}
	output := TxOutputs{
		Value:  100,
		PubKey: to,
	}
	transaction := Transaction{
		ID:      nil,
		Inputs:  []TxInputs{txInput},
		Outputs: []TxOutputs{output},
	}
	transaction.SetId()
	return &transaction
}
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

func (in *TxInputs) CanUnlock(data string) bool {
	return in.Sig == data
}
func (out *TxOutputs) CanBeUnlock(data string) bool {
	return out.PubKey == data
}

func NewTransaction(from, to string, amount int, chain *Blockchain) *Transaction {
	var inputs []TxInputs
	var outputs []TxOutputs
	acc, validOutputs := chain.FindSpendableOutputs(from, amount)
	if acc < amount {
		log.Panic("not enough funds")
	}
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		Handle(err)
		for _, item := range outs {
			input := TxInputs{
				ID:  txID,
				Out: item,
				Sig: from,
			}
			inputs = append(inputs, input)
		}
	}
	outputs = append(outputs, TxOutputs{
		Value:  amount,
		PubKey: to,
	})
	if acc > amount {
		outputs = append(outputs, TxOutputs{Value: acc - amount, PubKey: from})
	}
	tx := Transaction{ID: nil, Inputs: inputs, Outputs: outputs}
	tx.SetId()
	return &tx
}

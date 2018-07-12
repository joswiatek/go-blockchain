package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

// The reward for mining blocks
const subsidy = 10

// TxOutput represents the output of a transaction. It may be used as input.
type TxOutput struct {
	Value        int
	ScriptPubKey string
}

// TxInput represents the input of a transaction. It must reference an output.
type TxInput struct {
	TxId      []byte
	Vout      int
	ScriptSig string
}

// Transaction represents a single bitcoin transaction. It has an ID, a list
// inputs, and a list of outputs
type Transaction struct {
	ID   []byte
	Vin  []TxInput
	Vout []TxOutput
}

// CanUnlockOutputWith checks whether the input was initiated by the address
func (in *TxInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

// CanBeUnlockedWith checks if the output can be unlocked with the given data
func (out *TxOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}

// NewCoinbaseTx creates and returns a new coinbase transactions
func NewCoinbaseTx(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to %s", to)
	}

	txin := TxInput{
		TxId:      []byte{},
		Vout:      -1,
		ScriptSig: data,
	}
	txout := TxOutput{
		Value:        subsidy,
		ScriptPubKey: to,
	}

	tx := Transaction{nil, []TxInput{txin}, []TxOutput{txout}}
	tx.SetId()

	return &tx
}

func NewUnspentTxOutputTx(from, to string, amount int, bc *Blockchain) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	acc, validOutputs := bc.FindSpendableOutputs(from, amount)

	if acc < amount {
		log.Panic("ERROR: Not enough funds")
	}

	// Build a list of inputs for this transaction
	for txId, outs := range validOutputs {
		txIdStr, err := hex.DecodeString(txId)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := TxInput{
				TxId:      txIdStr,
				Vout:      out,
				ScriptSig: from}
			inputs = append(inputs, input)
		}
	}

	// Build a list of outputs

	// The requested transfer
	outputs = append(outputs, TxOutput{
		Value:        amount,
		ScriptPubKey: to,
	})

	// If there's any change left over
	if acc > amount {
		outputs = append(outputs, TxOutput{
			Value:        acc - amount,
			ScriptPubKey: from,
		})
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetId()

	return &tx
}

func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].TxId) == 0 && tx.Vin[0].Vout == -1
}

// SetId sets the ID of this transaction by hashing its encoding
func (tx *Transaction) SetId() {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

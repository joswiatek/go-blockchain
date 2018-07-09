package main

import "time"

// Block represents a single block in a blockchain.
type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

// NewBlock will create a new block given some data and the previous block's
// hash
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Data:          []byte(data),
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		Nonce:         0,
	}

	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// New NewGenesisBlock will create a new block that is meant to be the first
// block in a blockchain
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}

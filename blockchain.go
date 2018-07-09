package main

// Blockchain represents a single logical chain of blocks
type Blockchain struct {
	blocks []*Block
}

// AddBlock will add a block to the blockchain that contains the given data
func (bc *Blockchain) AddBlock(data string) {
	prevBlock := bc.blocks[len(bc.blocks)-1]
	newBlock := NewBlock(data, prevBlock.Hash)
	bc.blocks = append(bc.blocks, newBlock)
}

// NewBlockChain will create and return a new blockchain containining a single
// Genesis block
func NewBlockchain() *Blockchain {
	return &Blockchain{[]*Block{NewGenesisBlock()}}
}

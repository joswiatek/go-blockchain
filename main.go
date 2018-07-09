package main

import (
	"fmt"
	"strconv"
)

func main() {
	// Create a new blockchain
	bc := NewBlockchain()

	// Add some blocks. This is where the mining occurs, as an appropriate nonce
	// for the block must be found
	bc.AddBlock("Send 1 BTC to Joel")
	bc.AddBlock("Send 2 more BTC to Joel")

	// Display the blockchain
	for _, block := range bc.blocks {
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data:       %s\n", block.Data)
		fmt.Printf("Hash:       %x\n", block.Hash)

		// Validate that the appropriate work was done to add these blocks
		pow := NewProofOfWork(block)
		fmt.Printf("Pow:        %s\n", strconv.FormatBool(pow.Validate()))

		fmt.Println()
	}
}

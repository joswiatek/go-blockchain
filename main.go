package main

import "fmt"

func main() {
	// Create a new blockchain
	bc := NewBlockchain()

	// Add some blocks
	bc.AddBlock("Send 1 BTC to Joel")
	bc.AddBlock("Send 2 more BTC to Joel")

	// Display the blockchain
	for _, block := range bc.blocks {
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data:       %s\n", block.Data)
		fmt.Printf("Hash:       %x\n", block.Hash)
		fmt.Println()
	}
}

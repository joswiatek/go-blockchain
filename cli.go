package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

// CLI is used to manage command line interactions
type CLI struct {
	bc *Blockchain
}

// printUsage displays the usage instructions for the CLI
func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Printf("  %-30s%s\n", "addblock -data BLOCK_DATA", "add a block to the blockchain")
	fmt.Printf("  %-30s%s\n", "printchain ", "prints all the blocks of the block chain")
}

// validateArgs ensures that there are at least two arguments
func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

// Run will validate and parse the arguments, then perform the appropriate
// action
func (cli *CLI) Run() {
	cli.validateArgs()

	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	addBlockData := addBlockCmd.String("data", "", "Block data")

	switch os.Args[1] {
	case "addblock":
		err := addBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if addBlockCmd.Parsed() {
		// Verify that they provided data
		if *addBlockData == "" {
			addBlockCmd.Usage()
			os.Exit(1)
		}
		cli.addBlock(*addBlockData)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}
}

// addBlock simply adds a block the to the blockchain
func (cli *CLI) addBlock(data string) {
	cli.bc.AddBlock(data)
	fmt.Println("Success!")
}

// printChain uses a BlockchainIterator to print out the blockchain, from newest
// to oldest blocks
func (cli *CLI) printChain() {
	bci := cli.bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("%-10s : %x\n", "Prev. hash", block.PrevBlockHash)
		fmt.Printf("%-10s : %s\n", "Data", block.Data)
		fmt.Printf("%-10s : %x\n", "Hash", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("%-10s : %s\n", "PoW", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		// Check if we're at the genesis block
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

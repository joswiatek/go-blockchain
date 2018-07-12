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
}

// printUsage displays the usage instructions for the CLI
func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Printf("  %-30s%s\n", "createblockchain -address ADDRESS", "Creates a blockchain and sends genesis block reward to ADDRESS")
	fmt.Printf("  %-30s%s\n", "getbalance -address ADDRESS", "Get the balance of ADDRESS")
	fmt.Printf("  %-30s%s\n", "send -from FROM -to TO -amount AMOUNT", "Send AMOUNT of coins from FROM address to TO address")
	fmt.Printf("  %-30s%s\n", "printchain ", "Prints all the blocks of the block chain")
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

	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send the genesis block reward to")
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	sendFromAddress := sendCmd.String("from", "", "Source wallet address")
	sendToAddress := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
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

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress)
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}

	if sendCmd.Parsed() {
		if *sendFromAddress == "" || *sendToAddress == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}
		cli.send(*sendFromAddress, *sendToAddress, *sendAmount)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}
}

func (cli *CLI) createBlockchain(address string) {
	// TODO: modify CreateBlockchain to not return the open blockchain
	bc := CreateBlockchain(address)
	bc.db.Close()
	fmt.Println("Blockchain created.")
}

func (cli *CLI) getBalance(address string) {
	bc := NewBlockchain()
	defer bc.db.Close()

	balance := 0
	unspentTxOutputs := bc.FindUnspentTxOutputs(address)

	for _, out := range unspentTxOutputs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CLI) send(from, to string, amount int) {
	bc := NewBlockchain()
	defer bc.db.Close()

	tx := NewUnspentTxOutputTx(from, to, amount, bc)
	bc.MineBlock([]*Transaction{tx})
	fmt.Println("Success!")
}

// printChain uses a BlockchainIterator to print out the blockchain, from newest
// to oldest blocks
func (cli *CLI) printChain() {
	bc := NewBlockchain()
	defer bc.db.Close()

	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("%-10s : %x\n", "Prev. hash", block.PrevBlockHash)
		fmt.Printf("%-10s : %v\n", "Transactions", block.Transactions)
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

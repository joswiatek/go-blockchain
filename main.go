package main

func main() {
	// Create a new blockchain
	bc := NewBlockchain()
	defer bc.db.Close()

	// Create and run the CLI
	cli := CLI{bc}
	cli.Run()
}

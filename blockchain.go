package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

const dbFile = "blockchain.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "This thing is off the chain!"

// Blockchain represents a single logical chain of blocks
type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

// BlockchainIterator is used to iterate through a blockchain that is stored in
// a bolt DB
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

// CreateBlockchain will create a new blockchain by creating a bucket in the
// boltDB and creating a genesis block. The reward for the genesis block will be
// sent to the address parameter
func CreateBlockchain(address string) *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists, cannot create a new one.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		cbtx := NewCoinbaseTx(address, genesisCoinbaseData)
		genesisBlock := NewGenesisBlock(cbtx)

		bucket, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		err = bucket.Put(genesisBlock.Hash, genesisBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = bucket.Put([]byte("l"), genesisBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		tip = genesisBlock.Hash

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{
		tip: tip,
		db:  db,
	}

	return &bc
}

// TODO: rename this method to OpenBlockchain or LoadBlockchain
// NewBlockchain will open a connection to the boltDB, and look for a bucket
// names "blocks". It will either read the bucket or create it, and set the tip
// of the blockchain to be the last block in the chain
func NewBlockchain() *Blockchain {
	if !dbExists() {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte

	// Open the database
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	// Try to retrieve the blockchain
	err = db.Update(func(tx *bolt.Tx) error {
		// Get the bucket
		bucket := tx.Bucket([]byte(blocksBucket))
		// Retrieve the last hash
		tip = bucket.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{
		tip: tip,
		db:  db,
	}
	return &bc
}

// Iterator creates and returns a blockchain iterator that points to the tip of
// the blockchain, and will iterate "backwards" towards older blocks.
func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{
		currentHash: bc.tip,
		db:          bc.db,
	}
}

func (bc *Blockchain) MineBlock(transactions []*Transaction) {
	var previousHash []byte

	// Retrieve the last hash
	err := bc.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		previousHash = bucket.Get([]byte("l"))
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, previousHash)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		err := bucket.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = bucket.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}
		bc.tip = newBlock.Hash
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}

// FindUnspentTransactions traverses the blockchain to find a list of
// transactions that contain unspent outputs that can be unlocked by address
func (bc *Blockchain) FindUnspentTxs(address string) []Transaction {
	var unspentTxs []Transaction
	spentTxOutputs := make(map[string][]int)
	bci := bc.Iterator()

	// For every block in this blockchain
	for {
		block := bci.Next()

		// For every transaction in this block
		for _, tx := range block.Transactions {

			txIdStr := hex.EncodeToString(tx.ID)

		Outputs:
			// For every output in this transaction
			for outIdx, out := range tx.Vout {
				// Check if this output was already spent further up in the blockchain
				// (we are moving in the direction of older transactions!)
				if spentTxOutputs[txIdStr] != nil {
					for _, spentOut := range spentTxOutputs[txIdStr] {
						// This output was spent, so we don't necessarily want the transaction
						if spentOut == outIdx {
							// jump to Outputs to check the rest of the outputs (one of the
							// other ouputs might be unspent!)
							continue Outputs
						}
					}
				}

				// TODO: check this before checking if the transaction was spent?

				// this output is unspent, check if we can unlock it
				if out.CanBeUnlockedWith(address) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}

			// If it's not a coinbase transaction, save all the inputs as spent outputs
			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						inTxId := hex.EncodeToString(in.TxId)
						spentTxOutputs[inTxId] = append(spentTxOutputs[inTxId], in.Vout)
					}
				}
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTxs
}

func (bc *Blockchain) FindUnspentTxOutputs(address string) []TxOutput {
	var unspentTxOutputs []TxOutput
	unspentTxs := bc.FindUnspentTxs(address)

	for _, tx := range unspentTxs {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				unspentTxOutputs = append(unspentTxOutputs, out)
			}
		}
	}
	return unspentTxOutputs
}

func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTxs := bc.FindUnspentTxs(address)
	accumulated := 0

Work:
	for _, tx := range unspentTxs {
		txIdStr := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txIdStr] = append(unspentOutputs[txIdStr], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
}

// Next will retrieve and return the current block, advancing the iterator
// backwards to the previous, older block
func (bci *BlockchainIterator) Next() *Block {
	var encodedBlock []byte

	// Retrieve the encoded version of the current block from the db
	err := bci.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		encodedBlock = bucket.Get(bci.currentHash)

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	block := Deserialize(encodedBlock)

	// Update the current block to step through the blockchain
	bci.currentHash = block.PrevBlockHash

	return block
}

// dbExists returns true if the database exists, false if it does not
func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

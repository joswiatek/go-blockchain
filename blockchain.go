package main

import (
	"log"

	"github.com/boltdb/bolt"
)

const dbFile = "blockchain.db"
const blocksBucket = "blocks"

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

// AddBlock will create a new block that contains the given data, and add it to
// this blockchain
func (bc *Blockchain) AddBlock(data string) {
	// Create a new block using the previous last hash
	newBlock := NewBlock(data, bc.tip)

	// Add the new block into the chain
	err := bc.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		// Insert the block
		err := bucket.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		// Update the last hash
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

// NewBlockchain will open a connection to the boltDB, and look for a bucket
// names "blocks". It will either read the bucket or create it, and set the tip
// of the blockchain to be the last block in the chain
func NewBlockchain() *Blockchain {
	var tip []byte

	// Open the database
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	// Try to retrieve the blockchain
	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		// If the bucket doesn't exist...
		if bucket == nil {
			// ...create the bucket...
			bucket, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				log.Panic(err)
			}

			// ...and place a genesis block
			genesis := NewGenesisBlock()
			err = bucket.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}

			err = bucket.Put([]byte("l"), genesis.Hash)
			if err != nil {
				log.Panic(err)
			}

			tip = genesis.Hash
		} else {
			// If the bucket exists, just retrieve the last hash
			tip = bucket.Get([]byte("l"))
			if tip == nil {
				log.Panic("BoltDB exists but does not contain the key \"l\"")
			}
		}
		return nil
	})

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

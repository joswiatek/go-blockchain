package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

// The higher this number, the higher the difficulty. A block's hash must fit into 256-targetBits bits
const targetBits = 16

// Cap the max nonce to avoid roll overs
var maxNonce = math.MaxInt64

// ProofOfWork is used to prove that the appropriate work was done
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

// NewProofOfWork creates and returns a new ProofOfWork struct
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	return &ProofOfWork{b, target}
}

// prepareData concatenates all the data in a block, including the nonce, into a
// single byte slice which can then be hashed
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join([][]byte{
		pow.block.PrevBlockHash,
		pow.block.HashTransactions(),
		IntToHex(pow.block.Timestamp),
		IntToHex(int64(targetBits)),
		IntToHex(int64(nonce)),
	},
		[]byte{},
	)
	return data
}

// Run will perform the brute force operations of repeatedly hashing and
// incrementing nonce until an appropriate value is found. It returns the nonce
// and hash it found
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	// TODO: remove
	// fmt.Printf("Mining the block containing \"%v\"\n", pow.block.Transactions)
	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)
		hashInt.SetBytes(hash[:])
		if hashInt.Cmp(pow.target) == -1 {
			break
		}
		nonce++
	}
	fmt.Print("\n\n")

	return nonce, hash[:]
}

// Validate will returns a true or false value of whether this ProofOfWork
// struct is valid
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}

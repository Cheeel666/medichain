package blockchain

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"time"
)

// Block describes block structure
type Block struct {
	Timestamp     int64
	Data          []byte // TODO: replace with structure?
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

// NewBlock ...
func NewBlock(data string, prevBlockHash []byte) Block {
	block := Block{
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

// SetHash ...
func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
	hash := sha256.Sum256(headers)

	b.Hash = hash[:]
}

// NewGenesisBlock - create genesis block
func NewGenesisBlock() Block {
	return NewBlock("Genesis Block", []byte{})
}

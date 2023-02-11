package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

var (
	LastSentBlockchainLen *int
	LastRcvdBlockchainLen *int
	BlockChainImpl        Blockchain
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

// Blockchain keeps a sequence of Blocks
type Blockchain struct {
	blocks []Block
}

// AddBlock saves provided data as a block in the blockchain
func (bc *Blockchain) AddBlock(data string) {
	prevBlock := bc.blocks[len(bc.blocks)-1]
	newBlock := NewBlock(data, prevBlock.Hash)
	bc.blocks = append(bc.blocks, newBlock)
}

func (bc *Blockchain) AddValidatedBlock(block Block) {
	bc.blocks = append(bc.blocks, block)
}

// NewBlockchain creates a new Blockchain with genesis Block
func NewBlockchain() *Blockchain {
	return &Blockchain{[]Block{NewGenesisBlock()}}
}

func (bc *Blockchain) ValidateBlocks() {
	for _, block := range bc.blocks {
		log.Info().Msg(fmt.Sprintf("Prev. hash: %x\n", block.PrevBlockHash))
		log.Info().Msg(fmt.Sprintf("Data: %s\n", block.Data))
		log.Info().Msg(fmt.Sprintf("Hash: %x\n", block.Hash))
		pow := NewProofOfWork(block)
		log.Info().Msg(fmt.Sprintf("PoW: %s\n", strconv.FormatBool(pow.Validate())))
	}
}

func (bc *Blockchain) Length() int {
	return len(bc.blocks)
}

func (bc *Blockchain) UpdateBc(blocks []Block) {
	bc.blocks = blocks
}

func (b *Blockchain) GetLastBlockHash() []byte {
	return b.blocks[len(b.blocks)-1].Hash
}

package blockchain

import (
	"fmt"
	"strconv"

	"github.com/rs/zerolog/log"
)

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

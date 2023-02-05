package blockchain

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"strconv"
)

// Blockchain keeps a sequence of Blocks
type Blockchain struct {
	blocks []*Block
}

// AddBlock saves provided data as a block in the blockchain
func (bc *Blockchain) AddBlock(data string) {
	prevBlock := bc.blocks[len(bc.blocks)-1]
	newBlock := NewBlock(data, prevBlock.Hash)
	bc.blocks = append(bc.blocks, newBlock)
}

// NewBlockchain creates a new Blockchain with genesis Block
func NewBlockchain() *Blockchain {
	return &Blockchain{[]*Block{NewGenesisBlock()}}
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

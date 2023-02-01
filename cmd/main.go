package main

import (
	"8sem/diploma/medichain/internal/blockchain"
)

func main() {
	bc := blockchain.NewBlockchain()

	bc.AddBlock("Test block")
	bc.AddBlock("Test 1")
	bc.AddBlock("Test 2")

	bc.ValidateBlocks()
}

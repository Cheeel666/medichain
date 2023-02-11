package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
	"time"
)

func handleStream(stream network.Stream) {
	log.Info().Msg(fmt.Sprintf("new stream"))

	buf := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	go p2pReadData(buf)
	go p2pWriteData(buf)

}

func p2pReadData(buf *bufio.ReadWriter) {
	for {
		str, err := buf.ReadString('\n')
		if err != nil {
			//log.Fatal(err)
			log.Warn().Msg(fmt.Sprintf("failed to read buf:%v", err))
		}

		if str == "" {
			return
		}
		if str == "Exit\n" {
			continue
		}

		chain := make([]Block, 0)
		if err := json.Unmarshal([]byte(str), &chain); err != nil {
			log.Fatal().Err(err)
		}
		mutex.Lock()
		if len(chain) >= BlockChainImpl.Length() {
			BlockChainImpl.UpdateBc(chain)

			//written, err := writeToFile()
			//if err != nil {
			//	log.Fatal().Msg(fmt.Sprintf("failed to save bc to file: %v", err))
			//}
			//log.Info().Msg(fmt.Sprintf("written %d bytes", written))

			bytes, err := json.MarshalIndent(BlockChainImpl.blocks, "", "  ")
			if err != nil {
				log.Fatal().Err(err)
			}
			// Green console color: 	\x1b[32m
			// Reset console color: 	\x1b[0m
			// fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
			if BlockChainImpl.Length() > *LastRcvdBlockchainLen {
				// Green console color: 	\x1b[32m
				// Reset console color: 	\x1b[0m
				fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
				*LastRcvdBlockchainLen = BlockChainImpl.Length()
			}
		}
		mutex.Unlock()

	}
}

func p2pWriteData(buf *bufio.ReadWriter) {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			mutex.Lock()
			bytes, err := json.Marshal(BlockChainImpl.blocks)
			if err != nil {
				log.Warn().Err(err)
			}
			mutex.Unlock()

			mutex.Lock()

			bytesWrote, err := buf.WriteString(string(bytes) + "\n")
			if err != nil {
				log.Error().Err(err)
			}

			log.Info().Msg(fmt.Sprintf("written %d bytes", bytesWrote))

			err = buf.Flush()
			if err != nil {
				log.Error().Err(err)
			}
			if BlockChainImpl.Length() > *LastSentBlockchainLen {
				fmt.Sprintf("%s\n", string(bytes))
				*LastSentBlockchainLen = BlockChainImpl.Length()
			}
			mutex.Unlock()

		}
	}()

	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			log.Fatal().Err(err)
		}

		if sendData != "\n" {
			sendData = strings.Replace(sendData, "\n", "", -1) + " (From terminal)"
			newBlock := NewBlock(sendData, BlockChainImpl.GetLastBlockHash())

			pow := NewProofOfWork(newBlock)
			if pow.Validate() {
				mutex.Lock()
				BlockChainImpl.AddValidatedBlock(newBlock)
				mutex.Unlock()
			}

			bytes, err := json.Marshal(BlockChainImpl.blocks)
			if err != nil {
				log.Warn().Err(err)
			}

			spew.Dump(BlockChainImpl)

			mutex.Lock()
			bytesWrote, err := buf.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			if err != nil {
				log.Error().Err(err)
			}

			log.Info().Msg(fmt.Sprintf("written %d bytes", bytesWrote))

			err = buf.Flush()
			if err != nil {
				log.Error().Err(err)
			}
			*LastSentBlockchainLen = BlockChainImpl.Length()
			mutex.Unlock()
		}
	}
}

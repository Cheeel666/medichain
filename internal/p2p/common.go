package p2p

import (
	"bufio"
	"encoding/json"
	"fmt"
	"medichain/internal/blockchain"
	"medichain/internal/models"
	"sync"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/rs/zerolog/log"
)

type streamHandler struct {
	peerInfo models.PeerHost
	mutex    *sync.RWMutex
	bc       *blockchain.Blockchain
}

func NewStreamHandler(peerHost models.PeerHost, bc *blockchain.Blockchain) *streamHandler {
	return &streamHandler{
		peerInfo: peerHost,
		mutex:    &sync.RWMutex{},
		bc:       bc,
	}
}

func (s *streamHandler) HandleStream(ns network.Stream) {
	log.Info().Msg(fmt.Sprintf("new list of peers: %v", s.peerInfo.BaseHost.Peerstore().Peers()))

	buf := bufio.NewReadWriter(bufio.NewReader(ns), bufio.NewWriter(ns))
	go s.p2pReadData(buf)
	go s.p2pWriteData(buf)

}

func (s *streamHandler) p2pReadData(buf *bufio.ReadWriter) {
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

		chain := make([]*blockchain.Block, 0)
		if err := json.Unmarshal([]byte(str), &chain); err != nil {
			log.Fatal().Err(err)
		}
		s.mutex.Lock()
		if len(chain) >= s.bc.Length() {
			s.bc.UpdateBc(chain)

			// save2File()

			bytes, err := json.MarshalIndent(Blockchain, "", "  ")
			if err != nil {
				log.Fatal().Err(err)
			}
			// Green console color: 	\x1b[32m
			// Reset console color: 	\x1b[0m
			// fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
			if len(Blockchain) > LastRcvdBlockchainLen {
				// Green console color: 	\x1b[32m
				// Reset console color: 	\x1b[0m
				fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
				LastRcvdBlockchainLen = len(Blockchain)
			}
		}
		s.mutex.Unlock()

	}
}

func (s *streamHandler) p2pWriteData(buf *bufio.ReadWriter) {

}

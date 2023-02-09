package p2p

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"medichain/config"
	"medichain/internal/blockchain"
	"medichain/internal/models"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/rs/zerolog/log"
)

type streamHandler struct {
	peerInfo      *models.PeerHost
	mutex         *sync.RWMutex
	bc            blockchain.Blockchain
	lastReadBCLen int
	lastSendBCLen int
	fileDir       string

	StreamHandler network.Stream
}

func NewStreamHandler(peerHost *models.PeerHost, bc blockchain.Blockchain, cfg *config.Config) *streamHandler {
	s := &streamHandler{
		peerInfo: peerHost,
		mutex:    &sync.RWMutex{},
		bc:       bc,
		fileDir:  fmt.Sprintf("%s_blockchain_%d.txt", cfg.BlockchainDir, cfg.PeerPort),
	}

	// TODO: refactor shitcode below
	peerHost.BaseHost.NewStream()
	s.StreamHandler = network.StreamHandler(s.HandleStream)

	return s
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

		chain := make([]blockchain.Block, 0)
		if err := json.Unmarshal([]byte(str), &chain); err != nil {
			log.Fatal().Err(err)
		}
		s.mutex.Lock()
		if len(chain) >= s.bc.Length() {
			s.bc.UpdateBc(chain)

			written, err := s.writeToFile()
			if err != nil {
				log.Fatal().Msg(fmt.Sprintf("failed to save bc to file: %v", err))
			}

			log.Info().Msg(fmt.Sprintf("written %d bytes", written))
			bytes, err := json.MarshalIndent(s.bc, "", "  ")
			if err != nil {
				log.Fatal().Err(err)
			}
			// Green console color: 	\x1b[32m
			// Reset console color: 	\x1b[0m
			// fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
			if s.bc.Length() > s.lastReadBCLen {
				// Green console color: 	\x1b[32m
				// Reset console color: 	\x1b[0m
				fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
				s.lastReadBCLen = s.bc.Length()
			}
		}
		s.mutex.Unlock()

	}
}

func (s *streamHandler) p2pWriteData(buf *bufio.ReadWriter) {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			s.mutex.Lock()
			bytes, err := json.Marshal(s.bc)
			if err != nil {
				log.Warn().Err(err)
			}
			s.mutex.Unlock()

			s.mutex.Lock()

			bytesWrote, err := buf.WriteString(string(bytes) + "\n")
			if err != nil {
				log.Error().Err(err)
			}

			log.Info().Msg(fmt.Sprintf("written %d bytes", bytesWrote))

			err = buf.Flush()
			if err != nil {
				log.Error().Err(err)
			}
			if s.bc.Length() > s.lastSendBCLen {
				fmt.Sprintf("%s\n", string(bytes))
				s.lastSendBCLen = s.bc.Length()
			}
			s.mutex.Unlock()

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
			newBlock := blockchain.NewBlock(sendData, s.bc.GetLastBlockHash())

			pow := blockchain.NewProofOfWork(newBlock)
			if pow.Validate() {
				s.mutex.Lock()
				s.bc.AddValidatedBlock(newBlock)
				s.mutex.Unlock()
			}

			bytes, err := json.Marshal(s.bc)
			if err != nil {
				log.Warn().Err(err)
			}

			spew.Dump(s.bc)

			s.mutex.Lock()
			bytesWrote, err := buf.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			if err != nil {
				log.Error().Err(err)
			}

			log.Info().Msg(fmt.Sprintf("written %d bytes", bytesWrote))

			err = buf.Flush()
			if err != nil {
				log.Error().Err(err)
			}
			s.lastSendBCLen = s.bc.Length()
			s.mutex.Unlock()
		}
	}
}

func (s *streamHandler) writeToFile() (int, error) {
	f, err := os.Open(s.fileDir)
	if err != nil {
		return 0, err
	}
	bcData, err := json.Marshal(s.bc)
	if err != nil {
		return 0, nil
	}

	bytesWritten, err := f.Write(bcData)
	if err != nil {
		return 0, err
	}
	return bytesWritten, nil
}

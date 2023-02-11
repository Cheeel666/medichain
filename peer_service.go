package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/core/crypto"
	peer2 "github.com/libp2p/go-libp2p/core/peer"
	pstore "github.com/libp2p/go-libp2p/core/peerstore"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"io"
	mrand "math/rand"
	"medichain/clients"
	"medichain/clients/discovery"
	"medichain/config"
	"medichain/models"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	mutex       *sync.RWMutex
	peerProfile models.PeerProfile
	PeerGraph   *models.PeerGraph
	PeerHost    *models.PeerHost
)

type Service struct {
	Discovery clients.DiscoveryClient
}

func NewService(cfg *config.Config) *Service {
	return &Service{
		Discovery: discovery.NewDiscoveryClient(cfg),
	}
}

// InitP2P - initialize p2p instance
func InitP2P(cfg *config.Config, s *Service) error {
	// request current Peer port
	p, err := s.Discovery.RequestPort()
	if err != nil {
		return err
	}

	if p.PeerPort == 0 {
		return errors.New("failed to get peer port from discovery")
	}

	// request peers graph
	PeerGraph, err = s.Discovery.RequestP2PGraph()
	if err != nil {
		return err
	}
	log.Info().Msg(fmt.Sprintf("requested p2p graph: %v", PeerGraph))

	PeerHost, err = makeHost(p.PeerPort, cfg.PeerListenerSeed)
	if err != nil {
		return err
	}
	log.Info().Msg(fmt.Sprintf("initialized peer host: %v", PeerHost))

	// TODO: set stream: create handler
	cfg.PeerPort = p.PeerPort

	PeerHost.BaseHost.SetStreamHandler("/p2p/1.0.0", handleStream)

	log.Info().Msg(fmt.Sprintf("before connecting: %v", PeerHost.BaseHost.Peerstore().Peers()))
	connectP2P(PeerHost)
	enrollP2P(s.Discovery)
	log.Info().Msg(fmt.Sprintf("after connecting: %v", PeerHost.BaseHost.Peerstore().Peers()))

	return nil
}

func makeHost(port int, seed int64) (*models.PeerHost, error) {
	// rand seed. Use pseudorandom if zero
	var r io.Reader
	if seed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(seed))
	}

	privKey, _, err := crypto.GenerateRSAKeyPair(2048, r)
	if err != nil {
		return nil, err
	}

	addrOpt, identityOpt := libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/%s/tcp/%d", GetMyIP(), port)),
		libp2p.Identity(privKey)

	baseHost, err := libp2p.New(addrOpt, identityOpt)
	if err != nil {
		return nil, err
	}

	log.Info().Msg(fmt.Sprintf("created host %v with id %v(%v)", baseHost, baseHost.ID(), baseHost.ID().String()))

	multiAddress, err := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", baseHost.ID().String()))
	if err != nil {
		return nil, err
	}

	log.Info().Msg(fmt.Sprintf("multiadress created: %v", multiAddress.String()))

	addr := baseHost.Addrs()[0]
	fullAddr := addr.Encapsulate(multiAddress)

	return &models.PeerHost{
		Ma:          multiAddress,
		BaseHost:    baseHost,
		FullAddress: fullAddr.String(),
	}, nil
}

func connectP2P(peerHost *models.PeerHost) {
	peerProfile.ThisPeer = models.Peer{PeerAddress: peerHost.FullAddress}

	if len(PeerGraph.Graph) == 0 { // first node in the network
		log.Info().Msg("I'm first peer. Creating Genesis Block.")
		BlockChainImpl.blocks = append(BlockChainImpl.blocks, NewGenesisBlock())
		// save2File()
		spew.Dump(BlockChainImpl.blocks)
		log.Info().Msg("I'm first peer. Listening for connections.")
	} else {
		log.Info().Msg("Connecting to P2P network")
		log.Info().Msg(fmt.Sprintf("Cardinality of PeerGraph = %d", len(PeerGraph.Graph)))

		// make connection with peers[choice]
		choice := genRandInt(len(PeerGraph.Graph))
		log.Info().Msg(fmt.Sprintf("Connecting choice = %v", choice))

		peers := make([]string, 0, len(PeerGraph.Graph))
		for p, _ := range PeerGraph.Graph {
			peers = append(peers, p)
		}
		log.Info().Msg(fmt.Sprintf("Connecting to %v", peers[choice]))
		connect2Target(peers[choice])
		peerProfile.Neighbors = append(peerProfile.Neighbors, models.Peer{PeerAddress: peers[choice]})

		log.Info().Msg(fmt.Sprintf("peers[choice] =%v ", peers[choice]))

	}
}

func enrollP2P(client clients.DiscoveryClient) { // Enroll to the P2P Network by adding THIS peer with Bootstrapper
	log.Info().Msg("Enrolling in P2P network at Bootstrapper")

	response, err := client.EnrollP2P(peerProfile)
	if err != nil {
		log.Error().Err(err)
		return
	}

	log.Info().Msg(fmt.Sprintf("response:%v", response))
}

func connect2Target(newTarget string) {
	log.Info().Msg(fmt.Sprintf("Attempting to connect to %v", newTarget))
	// The following code extracts target's peer ID from the
	// given multiaddress
	ipfsaddr, err := ma.NewMultiaddr(newTarget)
	if err != nil {
		log.Fatal().Err(err)
	}
	log.Printf("ipfsaddr = ", ipfsaddr)
	log.Printf("Target = ", newTarget)

	pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		log.Fatal().Err(err)
	}
	log.Printf("pid = ", pid)
	log.Printf("ma.P_IPFS = ", ma.P_IPFS)

	peerid, err := peer.IDB58Decode(pid)
	if err != nil {
		log.Fatal().Err(err)
	}
	log.Info().Msg(fmt.Sprintf("peerid = %v", peerid))

	// Decapsulate the /ipfs/<peerID> part from the target
	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
	targetPeerAddr, _ := ma.NewMultiaddr(
		fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))
	targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)
	log.Printf("targetPeerAddr = ", targetPeerAddr)
	log.Printf("targetAddr = ", targetAddr)

	// We have a peer ID and a targetAddr so we add it to the peerstore
	// so LibP2P knows how to contact it
	PeerHost.BaseHost.Peerstore().AddAddr(peer2.ID(peerid), targetAddr, pstore.PermanentAddrTTL)

	log.Info().Msg(fmt.Sprintf("opening stream to %v", newTarget))
	// make a new stream from host B to host A
	// it should be handled on host A by the handler we set above because
	// we use the same /p2p/1.0.0 protocol
	s, err := PeerHost.BaseHost.NewStream(context.Background(), peer2.ID(peerid), "/p2p/1.0.0")
	if err != nil {
		log.Fatal().Err(err)
	}
	// Create a buffered stream so that read and writes are non blocking.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	// Create a thread to read and write data.
	go p2pWriteData(rw)
	go p2pReadData(rw)

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-ch
		log.Info().Msg("Received Interrupt. Exiting now.")
		cleanup(rw)
		os.Exit(1)
	}()
	//select {} // hang forever
}

func cleanup(rw *bufio.ReadWriter) {
	fmt.Println("cleanup")
	mutex.Lock()
	rw.WriteString("Exit\n")
	rw.Flush()
	mutex.Unlock()
}

package service

import (
	"crypto/rand"
	"io"
	mrand "math/rand"
	"medichain/config"
	"medichain/internal/clients"
	"medichain/internal/clients/discovery"
	"medichain/internal/models"

	crypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/pkg/errors"
)

type Services struct {
	Discovery clients.DiscoveryClient
}

func NewService(cfg *config.Config) *Services {

	return &Services{
		Discovery: discovery.NewDiscoveryClient(cfg),
	}
}

func (s *Services) InitP2P(cfg config.Config) (*models.PeerProfile, error) {
	// request current Peer port
	peer, err := s.Discovery.RequestPort()
	if err != nil {
		return nil, err
	}

	if peer.PeerPort == 0 {
		return nil, errors.New("failed to get peer port from discovery")
	}

	// request peers graph
	p2pGraph, err := s.Discovery.RequestP2PGraph()
	if err != nil {
		return nil, err
	}

	makeHost(peer.PeerPort, cfg.PeerListenerSeed)
	//TODO: listening multiadress host

	return peer, nil
}

func makeHost(port int, seed int64) {
	// rand seed. Use pseudorandom if zero
	var r io.Reader
	if seed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(seed))
	}

	privKey, pubKey, err := crypto.GenerateRSAKeyPair(2048, r)
	if err != nil {
		return nil, err
	}

}

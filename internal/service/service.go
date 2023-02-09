package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	mrand "math/rand"
	"medichain/config"
	"medichain/internal/clients"
	"medichain/internal/clients/discovery"
	"medichain/internal/models"
	"medichain/internal/utils"

	"github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p/core/crypto"
	multiaddr "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type Services struct {
	Discovery clients.DiscoveryClient
}

func NewService(cfg *config.Config) *Services {

	return &Services{
		Discovery: discovery.NewDiscoveryClient(cfg),
	}
}

// TODO: create struct field for this

// InitP2P - initialize p2p instance
func (s *Services) InitP2P(ctx context.Context, cfg *config.Config) (*models.PeerProfile, error) {
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
	log.Info().Msg(fmt.Sprintf("requested p2p graph: %v", p2pGraph))

	peerHost, err := makeHost(ctx, peer.PeerPort, cfg.PeerListenerSeed)
	if err != nil {
		return nil, err
	}
	log.Info().Msg(fmt.Sprintf("initialized peer host: %v", peerHost))

	// TODO: set stream: create handler
	peerHost.BaseHost.SetStreamHandler()
	// TODO: connect p2p

	// TODO: enroll p2p

	return peer, nil
}

func makeHost(ctx context.Context, port int, seed int64) (*models.PeerHost, error) {
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

	addrOpt := libp2p.Option(libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/%s/tcp/%d", utils.GetMyIP(), port)))
	identityOpt := libp2p.Option(libp2p.Identity(privKey))

	baseHost, err := libp2p.New(addrOpt, identityOpt)
	if err != nil {
		return nil, err
	}

	log.Info().Msg(fmt.Sprintf("created host %v with id %v(%v)", baseHost, baseHost.ID(), baseHost.ID().String()))

	multiAddress, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ipfs/%s", baseHost.ID().String()))
	if err != nil {
		return nil, err
	}

	log.Info().Msg(fmt.Sprintf("multiadress created: %v", multiAddress.String()))

	addr := baseHost.Addrs()[0]
	fullAddr := addr.Encapsulate(multiAddress)

	peerFullAddress := fullAddr.String()
	return &models.PeerHost{
		Ma:          multiAddress,
		BaseHost:    baseHost,
		FullAddress: peerFullAddress,
	}, nil
}

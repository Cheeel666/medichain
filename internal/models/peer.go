package models

import (
	"sync"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/multiformats/go-multiaddr"
)

type Peer struct {
	PeerAddress string `json:"PeerAddress"`
}

// PeerProfile represents peer structure
type PeerProfile struct { // connections of one peer
	ThisPeer  Peer   `json:"ThisPeer"`  // any node
	PeerPort  int    `json:"PeerPort"`  // port of peer
	Neighbors []Peer `json:"Neighbors"` // edges to that node
	Status    bool   `json:"Status"`    // Status: Alive or Dead
	Connected bool   `json:"Connected"` // If a node is connected or not [To be used later]
}

// PeerGraph struct: represent structure of peers
type PeerGraph struct {
	Graph map[string]PeerProfile
	Mutex *sync.RWMutex
}

// PeerHost contains information about peer hosting
type PeerHost struct {
	Ma          multiaddr.Multiaddr
	BaseHost    host.Host
	FullAddress string
}

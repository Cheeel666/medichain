package models

import "sync"

type Peer struct {
	PeerAddress string `json:"PeerAddress"`
}

type PeerProfile struct { // connections of one peer
	ThisPeer  Peer   `json:"ThisPeer"`  // any node
	PeerPort  int    `json:"PeerPort"`  // port of peer
	Neighbors []Peer `json:"Neighbors"` // edges to that node
	Status    bool   `json:"Status"`    // Status: Alive or Dead
	Connected bool   `json:"Connected"` // If a node is connected or not [To be used later]
}

type PeerGraph struct {
	Graph map[string]PeerProfile
	Mutex *sync.RWMutex
}

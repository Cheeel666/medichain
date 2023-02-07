package clients

import "medichain/internal/models"

type DiscoveryClient interface {
	RequestPort() (*models.PeerProfile, error)
	RequestP2PGraph() (*models.PeerGraph, error)
}

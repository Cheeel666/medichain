package clients

import "medichain/internal/models"

type DiscoveryClient interface {
	RequestPort() (*models.PeerProfile, error)
	RequestP2PGraph() (*models.PeerGraph, error)
	EnrollP2P(profile models.PeerProfile) (*models.PeerProfile, error)
}

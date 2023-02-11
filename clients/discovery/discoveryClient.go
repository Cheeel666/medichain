package discovery

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"medichain/config"

	"medichain/clients"
	"medichain/models"
	"net/http"
	"sync"

	"github.com/rs/zerolog/log"
)

type discoveryClient struct {
	address string
}

func NewDiscoveryClient(cfg *config.Config) clients.DiscoveryClient {
	return &discoveryClient{
		address: cfg.DiscoveryAddress + cfg.DiscoveryPort,
	}
}

// RequestPort for peer
func (d *discoveryClient) RequestPort() (*models.PeerProfile, error) { // Requesting PeerPort
	peerProfile := &models.PeerProfile{}

	response, err := http.Get(fmt.Sprintf("http://%s/api/v1/request_port", d.address))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(responseData))
	err = json.Unmarshal(responseData, &peerProfile.PeerPort)
	if err != nil {
		return nil, err
	}

	log.Printf("Got response from peer discovery:%v", string(responseData))
	return peerProfile, nil
}

// RequestP2PGraph requests graph of p2p network
func (d *discoveryClient) RequestP2PGraph() (*models.PeerGraph, error) {
	peerGraph := &models.PeerGraph{
		Graph: make(map[string]models.PeerProfile),
		Mutex: &sync.RWMutex{},
	}
	response, err := http.Get(fmt.Sprintf("http://%s/api/v1/p2p_graph", d.address))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	peerGraph.Mutex.Lock()
	defer peerGraph.Mutex.Unlock()
	err = json.Unmarshal(responseData, peerGraph)
	if err != nil {
		return nil, err
	}

	return peerGraph, nil
}

// EnrollP2P adds peer to p2p network
func (d *discoveryClient) EnrollP2P(profile models.PeerProfile) (*models.PeerProfile, error) {
	peerProfile := &models.PeerProfile{}
	body, err := json.Marshal(profile)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("http://%s/api/v1/enroll_p2p", d.address)
	response, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	res, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(res, peerProfile)
	if err != nil {
		return nil, err
	}

	return peerProfile, nil
}

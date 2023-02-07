package p2p

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"medichain/config"
	"medichain/internal/models"
	"net/http"
)

func InitP2P(cfg *config.Config) (*models.PeerProfile, error) {
	// request current Peer port
	peer, err := requestPort(cfg.DiscoveryAddress + cfg.DiscoveryPort)
	if err != nil {
		return nil, err
	}

	if peer.PeerPort == 0 {
		return nil, errors.New("failed to get peer port from discovery")
	}

	//TODO: request peers graph

	//TODO: listening multiadress host

	return peer, nil
}

func requestPort(discoveryAddress string) (*models.PeerProfile, error) { // Requesting PeerPort
	peerProfile := &models.PeerProfile{}

	response, err := http.Get(fmt.Sprintf("http://%s/api/v1/request_port", discoveryAddress))
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

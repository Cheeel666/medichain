package p2p

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"medichain/config"
	"medichain/internal/models"
	"net/http"
)

func InitP2P(cfg *config.Config) (*models.PeerProfile, error) {
	// request current Peer port
	peer, err := requestPort(cfg.DiscoveryAddress)
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
	var peerProfile *models.PeerProfile

	peersReqAddr := discoveryAddress + "/peers"
	log.Println("call requestPort", peersReqAddr)

	response, err := http.Get(peersReqAddr)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(responseData, &peerProfile.PeerPort)
	if err != nil {
		return nil, err
	}

	log.Printf("Got response from peer discovery:%v", responseData)
	return peerProfile, nil
}

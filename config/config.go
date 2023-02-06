package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	DBHost           string `json:"db_host"`
	LogLevel         string `json:"log_level"`
	DiscoveryAddress string `json:"discovery_address"`
}

func InitConfig(path string) (*Config, error) {
	cfg := &Config{}

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(file, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

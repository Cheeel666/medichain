package main

import (
	"8sem/diploma/medichain/config"
	"8sem/diploma/medichain/internal/blockchain"
	"fmt"
	"github.com/rs/zerolog"

	"os"
)

const configPath = "config/config.json"

func main() {
	cfg, err := config.InitConfig(configPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	var level zerolog.Level
	switch cfg.LogLevel {
	case "debug":
		level = zerolog.DebugLevel
	case "info":
		level = zerolog.InfoLevel
	default:
		level = zerolog.WarnLevel
	}

	log := zerolog.New(os.Stdout).With().Timestamp().Logger().Level(level)
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return fmt.Sprintf(" %s:%d ", file, line)
	}

	log.Info().Msg("initialized config; starting app")

	bc := blockchain.NewBlockchain()

	bc.AddBlock("Test block")
	bc.AddBlock("Test 1")
	bc.AddBlock("Test 2")

	bc.ValidateBlocks()
}

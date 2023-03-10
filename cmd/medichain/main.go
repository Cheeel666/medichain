package main

import (
	"context"
	"fmt"
	"medichain/config"
	"medichain/internal/service"

	"github.com/rs/zerolog"

	"os"
)

const configPath = "config/config.json"

func main() {
	ctx := context.Background()

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

	svc := service.NewService(cfg)

	log.Info().Msg("initialized service; starting peer")

	// TODO: change returning value
	_, err = svc.InitP2P(ctx, cfg)
	if err != nil {
		log.Fatal().Msg(fmt.Sprintf("failed to init peer listener:%v", err))
	}

	// TODO: connect BC

	//
	//bc := blockchain.NewBlockchain()
	//
	//bc.AddBlock("Test block")
	//bc.AddBlock("Test 1")
	//bc.AddBlock("Test 2")
	//
	//bc.ValidateBlocks()
}

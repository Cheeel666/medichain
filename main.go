package main

import (
	"fmt"
	"github.com/rs/zerolog"
	"medichain/config"
	"os"
)

const configPath = "config/config.json"

func main() {
	//ctx := context.Background()

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
	log.Info().Msg(fmt.Sprintf("Parsed config: %v", cfg))

	svc := NewService(cfg)

	log.Info().Msg("initialized service; starting peer")

	// TODO: change returning value
	err = InitP2P(cfg, svc)
	if err != nil {
		log.Fatal().Msg(fmt.Sprintf("failed to init peer listener:%v", err))
	}

}

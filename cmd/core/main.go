package main

import (
	"github.com/notweuz/rcon-scheduler-core/internal/config"
	"github.com/notweuz/rcon-scheduler-core/internal/logger"

	"github.com/rs/zerolog/log"
)

func main() {
	logger.Setup()

	cfg, err := config.Load("core")
	if err != nil {
		log.Warn().Err(err).Msg("failed to load config")
	}
	log.Info().Msg("config loaded")

	logger.ChangeLevel(cfg.LogLevel)
}

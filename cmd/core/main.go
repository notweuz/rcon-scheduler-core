package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/notweuz/rcon-scheduler-core/internal/config"
	"github.com/notweuz/rcon-scheduler-core/internal/logger"
	"github.com/notweuz/rcon-scheduler-core/scheduler"
)

func main() {
	logger.Setup()

	cfg, err := config.Load("core")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}
	if err := logger.ChangeLevel(cfg.LogLevel); err != nil {
		log.Warn().Err(err).Msg("invalid log level")
	}

	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		log.Fatal().Err(err).Str("timezone", cfg.Timezone).Msg("invalid timezone")
	}

	entries, err := cfg.Build()
	if err != nil {
		log.Fatal().Err(err).Msg("invalid config")
	}

	sched, err := scheduler.New(
		scheduler.WithLogger(log.Logger),
		scheduler.WithLocation(loc),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create scheduler")
	}

	for _, e := range entries {
		if err := sched.Register(e.Server); err != nil {
			log.Fatal().Err(err).Msg("failed to register server")
		}
		for _, j := range e.Jobs {
			if _, err := sched.AddJob(e.Server.Key(), j); err != nil {
				log.Fatal().Err(err).Msg("failed to add job")
			}
		}
	}

	sched.Start()
	log.Info().Int("servers", len(entries)).Str("timezone", loc.String()).Msg("scheduler started")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	log.Info().Msg("shutting down")
	if err := sched.Shutdown(); err != nil {
		log.Error().Err(err).Msg("shutdown failed")
	}
}

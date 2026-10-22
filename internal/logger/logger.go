package logger

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

func Setup() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).With().Caller().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func ChangeLevel(new string) error {
	level, err := zerolog.ParseLevel(new)
	if err != nil {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		return fmt.Errorf("failed to parse log level, fallback to info: %w", err)
	}
	zerolog.SetGlobalLevel(level)
	log.Info().Any("lvl", level).Msg("log level changed")
	return nil
}

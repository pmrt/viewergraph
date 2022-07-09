package config

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

func setupLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}

package main

import (
	"github.com/pmrt/viewergraph/config"
	"github.com/pmrt/viewergraph/database"
	l "github.com/rs/zerolog/log"
)

func main() {
	l := l.With().
		Str("context", "app").
		Logger()

	l.Info().Msg("starting server")

	l.Info().Msg("setting up database connection")

	database.New()
}

func init() {
	config.Setup()
}

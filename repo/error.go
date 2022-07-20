package repo

import (
	l "github.com/rs/zerolog/log"
)

func handleErr(err error) {
	l.Error().Err(err).
		Str("context", "query").
		Msg("error while executing query")
}

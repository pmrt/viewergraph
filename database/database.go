package database

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"time"

	"github.com/pmrt/viewergraph/config"
	l "github.com/rs/zerolog/log"
)

type Storage interface {
	Ping(ctx context.Context) error
	Migrate() error
	Conn() *sql.DB
	Opts() *StorageOptions
}

type StorageOptions struct {
	StorageHost     string
	StoragePort     string
	StorageUser     string
	StoragePassword string
	StorageDbName   string

	StorageMaxIdleConns    int
	StorageMaxOpenConns    int
	StorageConnMaxLifetime time.Duration
	StorageConnTimeout     time.Duration

	MigrationVersion int
	MigrationPath    string
	DebugMode        bool
}

func New(sto Storage) Storage {
	l := l.With().
		Str("context", "database").
		Logger()
	opts := sto.Opts()

	l.Info().
		Str("host", opts.StorageHost).
		Str("port", opts.StoragePort).
		Str("db", opts.StorageDbName).
		Str("user", opts.StorageUser).
		Msg("=> => pinging database")
	ctx, cancel := context.WithTimeout(
		context.Background(),
		opts.StorageConnTimeout,
	)
	defer cancel()
	if err := sto.Ping(ctx); err != nil {
		l.Panic().Err(err).Msg("")
	}
	l.Info().Msg("=> => connection successful")

	if config.SkipMigrations {
		l.Info().Msg("=> => skipping migrations")
		return sto
	}

	l.Info().
		Int("mig_version", opts.MigrationVersion).
		Str("mig_path", opts.MigrationPath).
		Msg("=> => attempting to apply migrations")
	if err := sto.Migrate(); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			l.Panic().Err(err).Msg("")
		}
		l.Info().Msg("=> => => no changes were made")
	} else {
		l.Info().Msg("=> => => migration success")
	}

	return sto
}

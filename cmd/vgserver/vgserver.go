package main

import (
	"time"

	cfg "github.com/pmrt/viewergraph/config"
	"github.com/pmrt/viewergraph/database"
	"github.com/pmrt/viewergraph/database/clickhouse"
	"github.com/pmrt/viewergraph/database/postgres"
	l "github.com/rs/zerolog/log"
)

func main() {
	l := l.With().
		Str("context", "app").
		Logger()

	l.Info().Msg("starting server")

	l.Info().Msg("setting up database connection")

	l.Info().Msg("=> setting up clickhouse")
	database.New(clickhouse.New(
		&database.StorageOptions{
			StorageHost:     cfg.ClickhouseHost,
			StoragePort:     cfg.ClickhousePort,
			StorageUser:     cfg.ClickhouseUser,
			StoragePassword: cfg.ClickhousePassword,
			StorageDbName:   cfg.ClickhouseDBName,

			StorageMaxIdleConns:    cfg.ClickhouseMaxIdleConns,
			StorageMaxOpenConns:    cfg.ClickhouseMaxOpenConns,
			StorageConnMaxLifetime: time.Duration(cfg.ClickhouseConnMaxLifetimeMinutes) * time.Minute,
			StorageConnTimeout:     time.Duration(cfg.ClickhouseConnTimeoutSeconds) * time.Second,

			MigrationVersion: cfg.ClickhouseMigVersion,
			MigrationPath:    cfg.ClickhouseMigPath,

			DebugMode: cfg.Debug,
		}))
	l.Info().Msg("=> setting up postgres")
	database.New(postgres.New(
		&database.StorageOptions{
			StorageHost:     cfg.PostgresHost,
			StoragePort:     cfg.PostgresPort,
			StorageUser:     cfg.PostgresUser,
			StoragePassword: cfg.PostgresPassword,
			StorageDbName:   cfg.PostgresDBName,

			StorageMaxIdleConns:    cfg.PostgresMaxIdleConns,
			StorageMaxOpenConns:    cfg.PostgresMaxOpenConns,
			StorageConnMaxLifetime: time.Duration(cfg.PostgresConnMaxLifetimeMinutes) * time.Minute,
			StorageConnTimeout:     time.Duration(cfg.PostgresConnTimeoutSeconds) * time.Second,

			MigrationVersion: cfg.PostgresMigVersion,
			MigrationPath:    cfg.PostgresMigPath,
		}))
}

func init() {
	cfg.Setup()
}

package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/pmrt/viewergraph/config"
	l "github.com/rs/zerolog/log"
)

func ping(ctx context.Context, db *sql.DB) (err error) {
	timer := time.NewTicker(time.Second)
	for {
		select {
		case <-timer.C:
			if err = db.Ping(); err == nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func New() *sql.DB {
	l := l.With().
		Str("context", "database").
		Logger()

	l.Info().Msg("=> connecting to database")
	l.Debug().Msg("=> validating database connection")

	db := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{config.DBHost + ":" + config.DBPort},
		Auth: clickhouse.Auth{
			Database: config.DBName,
			Username: config.DBUser,
			Password: config.DBPass,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: time.Duration(config.DBDialTimeoutSeconds) * time.Second,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		Debug: config.Debug,
	})
	db.SetMaxIdleConns(config.DBMaxIdleConns)
	db.SetMaxOpenConns(config.DBMaxOpenConns)
	db.SetConnMaxLifetime(time.Hour)

	l.Info().Msg("=> pinging database")
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(config.DBDialTimeoutSeconds)*time.Second)
	defer cancel()
	if err := ping(ctx, db); err != nil {
		l.Panic().Err(err).Msg("")
	}

	l.Info().Msg("=> connection successful")
	return db
}

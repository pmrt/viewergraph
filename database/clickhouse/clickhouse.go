package clickhouse

import (
	"context"
	"database/sql"
	"time"

	ch "github.com/ClickHouse/clickhouse-go/v2"
	m "github.com/golang-migrate/migrate/v4"
	mch "github.com/golang-migrate/migrate/v4/database/clickhouse"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/pmrt/viewergraph/database"
)

type Clickhouse struct {
	db   *sql.DB
	opts *database.StorageOptions
}

func (s *Clickhouse) Ping(ctx context.Context) (err error) {
	timer := time.NewTicker(time.Second)
	for {
		select {
		case <-timer.C:
			if err = s.db.Ping(); err == nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *Clickhouse) Migrate() error {
	d, err := mch.WithInstance(s.db, &mch.Config{})
	if err != nil {
		return err
	}

	mg, err := m.NewWithDatabaseInstance(
		"file://"+s.opts.MigrationPath, "clickhouse", d,
	)
	if err != nil {
		return err
	}

	return mg.Migrate(uint(s.opts.MigrationVersion))
}

func (s *Clickhouse) Conn() *sql.DB {
	return s.db
}

func (s *Clickhouse) Opts() *database.StorageOptions {
	return s.opts
}

func New(opts *database.StorageOptions) database.Storage {
	db := ch.OpenDB(&ch.Options{
		Addr: []string{opts.StorageHost + ":" + opts.StoragePort},
		Auth: ch.Auth{
			Database: opts.StorageDbName,
			Username: opts.StorageUser,
			Password: opts.StoragePassword,
		},
		Settings: ch.Settings{
			"max_execution_time": 60,
		},
		Compression: &ch.Compression{
			Method: ch.CompressionLZ4,
		},
		Debug: opts.DebugMode,
	})
	db.SetMaxIdleConns(opts.StorageMaxIdleConns)
	db.SetMaxOpenConns(opts.StorageMaxOpenConns)
	db.SetConnMaxLifetime(opts.StorageConnMaxLifetime)

	return &Clickhouse{
		db:   db,
		opts: opts,
	}
}

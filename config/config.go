package config

import (
	"os"
	"reflect"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	l "github.com/rs/zerolog/log"
)

const LastMigrationVersion = 1

func Setup() {
	LoadVars()
	setupLogger()
}

type SupportStringconv interface {
	~int | ~int8 | ~int64 | ~float32 | ~string | ~bool
}

func conv(v string, to reflect.Kind) any {
	var err error

	if to == reflect.String {
		return v
	}

	if to == reflect.Bool {
		if bool, err := strconv.ParseBool(v); err == nil {
			return bool
		}
	}

	if to == reflect.Int {
		if int, err := strconv.Atoi(v); err == nil {
			return int
		}
	}

	if to == reflect.Int8 {
		if i64, err := strconv.ParseInt(v, 10, 8); err == nil {
			return int8(i64)
		}
	}

	if to == reflect.Int64 {
		if i64, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i64
		}
	}

	if to == reflect.Float32 {
		if f32, err := strconv.ParseFloat(v, 32); err == nil {
			return f32
		}
	}

	l.Panic().
		Err(err).
		Str("context", "config").
		Msg("")
	return nil
}

func Env[T SupportStringconv](key string, def T) T {
	if v, ok := os.LookupEnv(key); ok {
		val := conv(v, reflect.TypeOf(def).Kind()).(T)
		l.Debug().
			Str("context", "config").
			Msgf("=> [%s]: %v", key, val)
		return val
	}
	return def
}

var (
	ClickhouseHost                   string
	ClickhousePort                   string
	ClickhouseUser                   string
	ClickhousePassword               string
	ClickhouseDBName                 string
	ClickhouseMaxIdleConns           int
	ClickhouseMaxOpenConns           int
	ClickhouseConnTimeoutSeconds     int
	ClickhouseConnMaxLifetimeMinutes int
	ClickhouseMigVersion             int
	ClickhouseMigPath                string

	PostgresHost                   string
	PostgresPort                   string
	PostgresUser                   string
	PostgresPassword               string
	PostgresDBName                 string
	PostgresMaxIdleConns           int
	PostgresMaxOpenConns           int
	PostgresConnTimeoutSeconds     int
	PostgresConnMaxLifetimeMinutes int
	PostgresMigVersion             int
	PostgresMigPath                string

	HelixClientID string
	HelixSecret   string

	SkipMigrations bool

	APIPort string

	TrackIntervalMinutes int

	Debug    bool
	LogLevel int8
)

func LoadVars() {
	l := l.With().
		Str("context", "config").
		Logger()

	if err := godotenv.Load(); err != nil {
		l.Panic().
			Err(err).
			Msg("couldn't load .env file")
	}

	l.Info().Msg("reading environment variables")

	ClickhouseHost = Env("CLICKHOUSE_HOST", "127.0.0.1")
	ClickhousePort = Env("CLICKHOUSE_PORT", "8123")
	ClickhouseUser = Env("CLICKHOUSE_USER", "default")
	ClickhousePassword = Env("CLICKHOUSE_PASSWORD", "")
	ClickhouseDBName = Env("CLICKHOUSE_DB_NAME", "default")
	ClickhouseMaxIdleConns = Env("CLICKHOUSE_MAX_IDLE_CONNS", 5)
	ClickhouseMaxOpenConns = Env("CLICKHOUSE_MAX_OPEN_CONNS", 10)
	ClickhouseConnMaxLifetimeMinutes = Env("CLICKHOUSE_CONN_MAX_LIFETIME_MINUTES", 60)
	ClickhouseConnTimeoutSeconds = Env("CLICKHOUSE_CONN_TIMEOUT_SECONDS", 60)
	ClickhouseMigVersion = Env("CLICKHOUSE_MIG_VERSION", 1)
	ClickhouseMigPath = Env("CLICKHOUSE_MIG_PATH", "database/clickhouse/migrations")

	PostgresHost = Env("POSTGRES_HOST", "127.0.0.1")
	PostgresPort = Env("POSTGRES_PORT", "5432")
	PostgresUser = Env("POSTGRES_USER", "vg")
	PostgresPassword = Env("POSTGRES_PASSWORD", "unsafepassword")
	PostgresDBName = Env("POSTGRES_DB_NAME", "vg")
	PostgresMaxIdleConns = Env("POSTGRES_MAX_IDLE_CONNS", 5)
	PostgresMaxOpenConns = Env("POSTGRES_MAX_OPEN_CONNS", 10)
	PostgresConnMaxLifetimeMinutes = Env("POSTGRES_CONN_MAX_LIFETIME_MINUTES", 60)
	PostgresConnTimeoutSeconds = Env("POSTGRES_CONN_TIMEOUT_SECONDS", 60)
	PostgresMigVersion = Env("POSTGRES_MIG_VERSION", 1)
	PostgresMigPath = Env("POSTGRES_MIG_VERSION", "database/postgres/migrations")

	HelixClientID = Env("HELIX_CLIENT_ID", "fake_client_id")
	HelixSecret = Env("HELIX_SECRET", "fake_secret")

	TrackIntervalMinutes = Env("TRACK_INTERVAL_MINUTES", 3600)

	SkipMigrations = Env("SKIP_MIGRATIONS", false)

	Debug = Env("DEBUG", false)
	LogLevel = Env("LOG_LEVEL", int8(zerolog.DebugLevel))
}

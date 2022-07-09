package config

import (
	"os"
	"reflect"
	"strconv"

	"github.com/joho/godotenv"
	l "github.com/rs/zerolog/log"
)

const LastMigrationVersion = 1

func Setup() {
	LoadVars()
	setupLogger()
}

type SupportStringconv interface {
	~int | ~int64 | ~float32 | ~string | ~bool
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
	DBHost               string
	DBUser               string
	DBPass               string
	DBPort               string
	DBName               string
	DBMaxIdleConns       int
	DBMaxOpenConns       int
	DBDialTimeoutSeconds int

	APIPort string

	Debug bool
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

	DBHost = Env("DB_HOST", "127.0.0.1")
	DBPort = Env("DB_PORT", "8123")
	DBUser = Env("DB_USER", "default")
	DBPass = Env("DB_PASS", "")
	DBName = Env("DB_NAME", "default")
	DBMaxIdleConns = Env("DB_MAX_IDLE_CONS", 5)
	DBMaxOpenConns = Env("DB_MAX_OPEN_CONS", 10)
	DBDialTimeoutSeconds = Env("DB_DIAL_TIMEOUT_SECONDS", 60)

	Debug = Env("DEBUG", false)

}

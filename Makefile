#!make
include .env

pgmigrations := $(wildcard ./database/postgres/migrations/*.sql)

.PHONY: tools test

all: gen

# generate sql builder files for type safe SQL
gen: $(pgmigrations)
	@echo Detected change in postgres migrations, generating new SQL types
	# remove entire dir so make knows it has been updated
	rm -rf /.gen
	jet -dsn=postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB_NAME}?sslmode=disable -schema=${POSTGRES_SCHEMA} -path=./gen

# migrate postgres 1 step down
pgdown:
	migrate -path database/postgres/migrations -database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB_NAME}?sslmode=disable down 1
# migrate postgres 1 step up
pgup:
	migrate -path ./database/postgres/migrations -database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB_NAME}?sslmode=disable up 1

tools:
	# Jet cli for generating types for SQL
	@echo Installing go-jet
	go install github.com/go-jet/jet/v2/cmd/jet@latest
	# TODO - maybe include go-migrate (check build constraints)

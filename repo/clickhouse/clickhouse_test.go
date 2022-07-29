package clickhouse

import (
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/pmrt/viewergraph/database"
	ch "github.com/pmrt/viewergraph/database/clickhouse"
)

var db *sql.DB

func TestMain(m *testing.M) {
	// Run a docker with a database for testing
	pool, err := dockertest.NewPool("")
	if err != nil {
		panic(err)
	}
	res, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "clickhouse/clickhouse-server",
		Env: []string{
			"CLICKHOUSE_DB=name",
			"CLICKHOUSE_USER=user",
			"CLICKHOUSE_PASSWORD=test",
			"CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT=1",
		},
	}, func(hc *docker.HostConfig) {
		hc.AutoRemove = true
		hc.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		panic(err)
	}
	res.Expire(120)

	// Prepare a connection to the db in the docker
	sto := database.New(
		ch.New(&database.StorageOptions{
			StorageHost:            res.GetBoundIP("9000/tcp"),
			StoragePort:            res.GetPort("9000/tcp"),
			StorageUser:            "user",
			StoragePassword:        "test",
			StorageDbName:          "name",
			StorageMaxIdleConns:    5,
			StorageMaxOpenConns:    10,
			StorageConnMaxLifetime: time.Hour,
			StorageConnTimeout:     60 * time.Second,
			DebugMode:              true,

			MigrationVersion: 1,
			MigrationPath:    "../../database/clickhouse/migrations",
		}))
	db = sto.Conn()

	// Run tests
	code := m.Run()

	if err := pool.Purge(res); err != nil {
		log.Fatal(err)
	}
	os.Exit(code)
}

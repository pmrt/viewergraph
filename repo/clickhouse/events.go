package clickhouse

import (
	"database/sql"
	"time"

	"github.com/pmrt/viewergraph/utils"
)

type Viewers struct {
	ts      time.Time
	Viewers []string
	Channel string
}

type RawEvent struct {
	Ts        time.Time
	Username  string
	Channel   string
	EventType string `db:"event_type"`
}

// INSERT INTO raw_events
// VALUES (toDateTime('2022-07-14 08:00:00'), 'user1', 'alexelcapo', 'view');
func InsertViewers(db *sql.DB, vw *Viewers) error {
	l := utils.Logger("query")

	tx, err := db.Begin()
	if err != nil {
		l.Error().Err(err).Msg("error while opening transaction")
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO raw_events (ts, username, channel, event_type)")
	if err != nil {
		l.Error().Err(err).Msg("error while preparing statement")
		return err
	}

	// we round time to start of hour for aggregation purposes. Hour is the
	// smallest unit we will be storing in the database.
	t := time.Date(vw.ts.Year(), vw.ts.Month(), vw.ts.Day(), vw.ts.Hour(), 0, 0, 0, vw.ts.Location())
	for _, usr := range vw.Viewers {
		if _, err := stmt.Exec(t, usr, vw.Channel, "view"); err != nil {
			l.Error().Err(err).Msg("error while adding values to the batch")
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		l.Error().Err(err).Msg("error while committing transaction")
		return err
	}
	return nil
}

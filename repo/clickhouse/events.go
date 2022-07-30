package clickhouse

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/davecgh/go-spew/spew"
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

type Event struct {
	Ts       time.Time
	Username string
	Channel  string
	Referrer string
}

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

func ReconcileEvents(db *sql.DB, lastAt time.Time, window time.Duration) error {
	l := utils.Logger("query")

	const margin = 15 * time.Minute
	// The window is the max. time an event can have relation with future events
	// (or future events with previous ones), covering all the events than can
	// have events related. Then we round it to the start of the our since events
	// are also rounded and we add an extra margin just in case.
	//
	// So that should select all the elements that can have elements related to
	// them and that are not already processed (plus others that are already
	// processed). Duplicates are handled by the database with a
	// ReplacingMergeTree.
	t := lastAt.Add(-window)
	t2 := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
	since := t2.Add(-margin)

	l.Info().Msgf("event reconciliation since: %s", since)

	row := db.QueryRow(`
    INSERT INTO events
    SELECT
      ts, username, channel,
      arrayJoin(referrers) as referrer
    FROM (
      SELECT
        ts, username, channel,
        groupArray(channel) OVER (
          PARTITION BY username
          ORDER BY
           ts ASC
          RANGE BETWEEN @WindowSeconds PRECEDING AND 0 PRECEDING
        ) AS referrers
      FROM raw_events
      WHERE
        event_type = 'view' AND
        ts >= @Since
    )
    WHERE
      length(referrers) > 0 AND referrer != channel
    ORDER BY (channel, ts, referrer)
  `,
		sql.Named("WindowSeconds", window.Seconds()),
		sql.Named("Since", since),
	)

	fmt.Print(spew.Sdump(row))
	return nil
}

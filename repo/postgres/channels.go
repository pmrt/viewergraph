package postgres

import (
	"database/sql"
	//lint:ignore ST1001 This library is prepared for dot imports
	. "github.com/go-jet/jet/v2/postgres"

	//lint:ignore ST1001 This library is prepared for dot imports
	"github.com/pmrt/viewergraph/gen/vg/public/model"
	. "github.com/pmrt/viewergraph/gen/vg/public/table"
	"github.com/pmrt/viewergraph/utils"
)

// Tracked retrieves the tracked channels ids from a `db` source
func Tracked(db *sql.DB) (f []*model.TrackedChannels, err error) {
	l := utils.Logger("query")

	stmt := SELECT(
		TrackedChannels.BroadcasterID,
	).FROM(TrackedChannels)

	if err = stmt.Query(db, &f); err != nil {
		l.Error().Err(err).Msg("error while executing query")
		return f, err
	}
	return f, nil
}

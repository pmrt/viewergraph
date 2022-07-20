package repo

import (
	"database/sql"
	//lint:ignore ST1001 This library is prepared for dot imports
	. "github.com/go-jet/jet/v2/postgres"
	//lint:ignore ST1001 This library is prepared for dot imports
	"github.com/pmrt/viewergraph/gen/vg/public/model"
	. "github.com/pmrt/viewergraph/gen/vg/public/table"
)

func Tracked(db *sql.DB) (f []*model.TrackedChannels, err error) {
	stmt := SELECT(
		TrackedChannels.BroadcasterID,
	).FROM(TrackedChannels)

	if err = stmt.Query(db, &f); err != nil {
		handleErr(err)
		return f, err
	}
	return f, nil
}

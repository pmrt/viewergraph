package repo

import (
	"log"
	"testing"

	"github.com/go-test/deep"
	"github.com/pmrt/viewergraph/gen/vg/public/model"
	"github.com/pmrt/viewergraph/utils"

	//lint:ignore ST1001 This library is prepared for dot imports
	. "github.com/pmrt/viewergraph/gen/vg/public/table"
)

func insertChannel(channel *model.TrackedChannels) {
	stmt := TrackedChannels.INSERT(
		TrackedChannels.AllColumns,
	).MODEL(channel)

	_, err := stmt.Exec(db)
	if err != nil {
		log.Fatal(err)
	}
}

func TestChannels(t *testing.T) {
	insertChannel(&model.TrackedChannels{
		BroadcasterID:          "36138196",
		BroadcasterDisplayName: "alexelcapo",
		BroadcasterUsername:    "alexelcapo",
		BroadcasterType:        model.Broadcastertype_Partner,
		ProfileImageURL:        utils.StrPtr("https://static-cdn.jtvnw.net/jtv_user_pictures/bf455aac-4ce9-4daa-94a0-c6c0a1b2500d-channel_offline_image-1920x1080.png"),
		OfflineImageURL:        utils.StrPtr("https://static-cdn.jtvnw.net/jtv_user_pictures/bf455aac-4ce9-4daa-94a0-c6c0a1b2500d-channel_offline_image-1920x1080.png"),
	})

	rows, err := Tracked(db)
	if err != nil {
		t.Fatal(err)
	}

	if len(rows) != 1 {
		t.Fatal("expected channels to return exactly 1 row")
	}

	want := &model.TrackedChannels{
		BroadcasterID: "36138196",
	}
	if diff := deep.Equal(rows[0], want); diff != nil {
		t.Fatal(diff)
	}
}

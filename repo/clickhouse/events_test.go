package clickhouse

import (
	"testing"
	"time"

	"github.com/go-test/deep"
)

func TestInsertViewers(t *testing.T) {
	ts, err := time.Parse(time.RFC3339, "2020-10-11T10:30:20.123Z")
	if err != nil {
		t.Fatal(err)
	}

	vw := &Viewers{
		ts:      ts,
		Viewers: []string{"user1", "user2", "user3", "user4", "user5"},
		Channel: "streamer1",
	}

	if err := InsertViewers(db, vw); err != nil {
		t.Fatal(err)
	}

	rows, err := db.Query("SELECT toTimeZone(ts, 'UTC'), username, channel, event_type FROM raw_events")
	if err != nil {
		t.Fatal(err)
	}
	got := make([]*RawEvent, 0, len(vw.Viewers))
	for rows.Next() {
		evt := new(RawEvent)
		if err := rows.Scan(
			&evt.Ts,
			&evt.Username,
			&evt.Channel,
			&evt.EventType,
		); err != nil {
			t.Fatal(err)
		}
		got = append(got, evt)
	}

	wantTs, err := time.Parse(time.RFC3339, "2020-10-11T10:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	want := []*RawEvent{
		{Ts: wantTs, Username: "user1", Channel: "streamer1", EventType: "view"},
		{Ts: wantTs, Username: "user2", Channel: "streamer1", EventType: "view"},
		{Ts: wantTs, Username: "user3", Channel: "streamer1", EventType: "view"},
		{Ts: wantTs, Username: "user4", Channel: "streamer1", EventType: "view"},
		{Ts: wantTs, Username: "user5", Channel: "streamer1", EventType: "view"},
	}
	if diff := deep.Equal(got, want); diff != nil {
		t.Fatal(diff)
	}
}

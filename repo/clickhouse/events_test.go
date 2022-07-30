package clickhouse

import (
	"database/sql"
	"io"
	"testing"
	"time"

	"github.com/go-test/deep"
)

func cleanTable(table string) {
	// Probably very unsafe and vulnerable to SQL Injections but it is just for
	// testing. We can't pass table names as parameters.
	_ = db.QueryRow("TRUNCATE TABLE " + table)
}

func parseTime(timestr string) time.Time {
	ts, err := time.Parse(time.RFC3339, timestr)
	if err != nil {
		panic(err)
	}
	return ts
}

func TestInsertViewers(t *testing.T) {
	t.Cleanup(func() {
		cleanTable("raw_events")
	})

	ts := parseTime("2020-10-11T10:30:20.123Z")

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

	wantTs := parseTime("2020-10-11T10:00:00Z")
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

func insertRawEvent(ts, username, channel, evttype string) {
	_ = db.QueryRow(
		"INSERT INTO raw_events VALUES (@Ts, @Username, @Channel, @EvtType)",
		sql.Named("Ts", parseTime(ts)),
		sql.Named("Username", username),
		sql.Named("Channel", channel),
		sql.Named("EvtType", evttype),
	)
}

func TestReconcileTimings(t *testing.T) {
	t.Cleanup(func() {
		cleanTable("raw_events")
		cleanTable("events")
		cleanTable("aggregated_flows_by_dst")
		cleanTable("aggregated_flows_by_src")
	})

	/*
	            10          11          12          13          15
	  events:   |                       |           |           |
	  recon.:                 |                       |                |
	*/

	insertRawEvent(
		"2020-10-11T10:00:00Z",
		"user1",
		"alexelcapo",
		"view",
	)
	if err := ReconcileEvents(db, time.Time{}, 2*time.Hour); err != nil {
		t.Fatal(err)
	}
	insertRawEvent(
		"2020-10-11T12:00:00Z",
		"user1",
		"jujalag",
		"view",
	)
	insertRawEvent(
		"2020-10-11T13:00:00Z",
		"user1",
		"chuso",
		"view",
	)
	if err := ReconcileEvents(db, parseTime("2020-10-11T11:05:00Z"), 2*time.Hour); err != nil {
		t.Fatal(err)
	}
	insertRawEvent(
		"2020-10-11T15:00:00Z",
		"user1",
		"yuste",
		"view",
	)
	if err := ReconcileEvents(db, parseTime("2020-10-11T13:05:00Z"), 2*time.Hour); err != nil {
		t.Fatal(err)
	}
	row := db.QueryRow("OPTIMIZE TABLE events")
	if err := row.Err(); err != nil {
		if err != io.EOF {
			t.Fatal(err)
		}
	}

	rows, err := db.Query("SELECT toTimeZone(ts, 'UTC'), username, channel, referrer FROM events ORDER BY ts")
	if err != nil {
		t.Fatal(err)
	}
	got := make([]*Event, 0, 3)
	for rows.Next() {
		evt := new(Event)
		if err := rows.Scan(
			&evt.Ts,
			&evt.Username,
			&evt.Channel,
			&evt.Referrer,
		); err != nil {
			t.Fatal(err)
		}
		got = append(got, evt)
	}

	want := []*Event{
		{Ts: parseTime("2020-10-11T12:00:00Z"), Username: "user1", Channel: "jujalag", Referrer: "alexelcapo"},
		{Ts: parseTime("2020-10-11T13:00:00Z"), Username: "user1", Channel: "chuso", Referrer: "jujalag"},
		{Ts: parseTime("2020-10-11T15:00:00Z"), Username: "user1", Channel: "yuste", Referrer: "chuso"},
	}
	if diff := deep.Equal(got, want); diff != nil {
		t.Fatal(diff)
	}
}

func TestReconcileSameTime(t *testing.T) {
	t.Cleanup(func() {
		cleanTable("raw_events")
		cleanTable("events")
		cleanTable("aggregated_flows_by_dst")
		cleanTable("aggregated_flows_by_src")
	})

	insertRawEvent(
		"2020-10-11T10:00:00Z",
		"user1",
		"alexelcapo",
		"view",
	)
	insertRawEvent(
		"2020-10-11T10:00:00Z",
		"user1",
		"jujalag",
		"view",
	)
	if err := ReconcileEvents(db, time.Time{}, 2*time.Hour); err != nil {
		t.Fatal(err)
	}
	row := db.QueryRow("OPTIMIZE TABLE events")
	if err := row.Err(); err != nil {
		if err != io.EOF {
			t.Fatal(err)
		}
	}

	rows, err := db.Query("SELECT toTimeZone(ts, 'UTC'), username, channel, referrer FROM events ORDER BY ts")
	if err != nil {
		t.Fatal(err)
	}
	got := make([]*Event, 0, 3)
	for rows.Next() {
		evt := new(Event)
		if err := rows.Scan(
			&evt.Ts,
			&evt.Username,
			&evt.Channel,
			&evt.Referrer,
		); err != nil {
			t.Fatal(err)
		}
		got = append(got, evt)
	}

	want := []*Event{
		{Ts: parseTime("2020-10-11T10:00:00Z"), Username: "user1", Channel: "alexelcapo", Referrer: "jujalag"},
		{Ts: parseTime("2020-10-11T10:00:00Z"), Username: "user1", Channel: "jujalag", Referrer: "alexelcapo"},
	}
	if diff := deep.Equal(got, want); diff != nil {
		t.Fatal(diff)
	}
}

func TestReconcileMultipleUsers(t *testing.T) {
	t.Cleanup(func() {
		cleanTable("raw_events")
		cleanTable("events")
		cleanTable("aggregated_flows_by_dst")
		cleanTable("aggregated_flows_by_src")
	})

	insertRawEvent(
		"2020-10-11T07:00:00Z",
		"user1",
		"alexelcapo",
		"view",
	)
	insertRawEvent(
		"2020-10-11T07:00:00Z",
		"user2",
		"alexelcapo",
		"view",
	)
	insertRawEvent(
		"2020-10-11T07:00:00Z",
		"user3",
		"alexelcapo",
		"view",
	)
	insertRawEvent(
		"2020-10-11T08:00:00Z",
		"user1",
		"jujalag",
		"view",
	)
	insertRawEvent(
		"2020-10-11T08:00:00Z",
		"user2",
		"jujalag",
		"view",
	)
	insertRawEvent(
		"2020-10-11T08:00:00Z",
		"user3",
		"chuso",
		"view",
	)

	if err := ReconcileEvents(db, time.Time{}, 2*time.Hour); err != nil {
		t.Fatal(err)
	}
	row := db.QueryRow("OPTIMIZE TABLE events")
	if err := row.Err(); err != nil {
		if err != io.EOF {
			t.Fatal(err)
		}
	}

	rows, err := db.Query("SELECT toTimeZone(ts, 'UTC'), username, channel, referrer FROM events ORDER BY username")
	if err != nil {
		t.Fatal(err)
	}
	got := make([]*Event, 0, 3)
	for rows.Next() {
		evt := new(Event)
		if err := rows.Scan(
			&evt.Ts,
			&evt.Username,
			&evt.Channel,
			&evt.Referrer,
		); err != nil {
			t.Fatal(err)
		}
		got = append(got, evt)
	}

	want := []*Event{
		{Ts: parseTime("2020-10-11T08:00:00Z"), Username: "user1", Channel: "jujalag", Referrer: "alexelcapo"},
		{Ts: parseTime("2020-10-11T08:00:00Z"), Username: "user2", Channel: "jujalag", Referrer: "alexelcapo"},
		{Ts: parseTime("2020-10-11T08:00:00Z"), Username: "user3", Channel: "chuso", Referrer: "alexelcapo"},
	}
	if diff := deep.Equal(got, want); diff != nil {
		t.Fatal(diff)
	}

}

func TestFlowsDstHourly(t *testing.T) {
	t.Cleanup(func() {
		cleanTable("raw_events")
		cleanTable("events")
		cleanTable("aggregated_flows_by_dst")
		cleanTable("aggregated_flows_by_src")
	})

	insertRawEvent(
		"2020-10-11T08:00:00Z",
		"user1",
		"jujalag",
		"view",
	)
	insertRawEvent(
		"2020-10-11T08:00:00Z",
		"user2",
		"jujalag",
		"view",
	)
	insertRawEvent(
		"2020-10-11T10:00:00Z",
		"user3",
		"felipez",
		"view",
	)
	insertRawEvent(
		"2020-10-11T08:00:00Z",
		"user4",
		"jujalag",
		"view",
	)
	insertRawEvent(
		"2020-10-11T09:00:00Z",
		"user5",
		"felipez",
		"view",
	)
	insertRawEvent(
		"2020-10-11T10:00:00Z",
		"user1",
		"alexelcapo",
		"view",
	)
	insertRawEvent(
		"2020-10-11T10:00:00Z",
		"user2",
		"alexelcapo",
		"view",
	)
	insertRawEvent(
		"2020-10-11T10:00:00Z",
		"user3",
		"alexelcapo",
		"view",
	)
	insertRawEvent(
		"2020-10-11T10:00:00Z",
		"user4",
		"alexelcapo",
		"view",
	)
	insertRawEvent(
		"2020-10-11T10:00:00Z",
		"user5",
		"alexelcapo",
		"view",
	)
	insertRawEvent(
		"2020-10-11T10:00:00Z",
		"user6",
		"yuste",
		"view",
	)
	insertRawEvent(
		"2020-10-11T12:00:00Z",
		"user6",
		"alexelcapo",
		"view",
	)
	if err := ReconcileEvents(db, time.Time{}, 2*time.Hour); err != nil {
		t.Fatal(err)
	}

	got, err := UserFlowsByDstHourly(
		db,
		"alexelcapo",
		parseTime("2020-10-11T08:00:00Z"),
		parseTime("2020-10-11T12:00:00Z"),
	)
	if err != nil {
		t.Fatal(err)
	}

	want := []*UserFlowDst{
		{Ts: parseTime("2020-10-11T10:00:00Z"), Referrer: "jujalag", Total: 3},
		{Ts: parseTime("2020-10-11T10:00:00Z"), Referrer: "felipez", Total: 2},
		{Ts: parseTime("2020-10-11T12:00:00Z"), Referrer: "yuste", Total: 1},
	}
	if diff := deep.Equal(got, want); diff != nil {
		t.Fatal(diff)
	}
}

func TestFlowsSrcHourly(t *testing.T) {
	t.Cleanup(func() {
		cleanTable("raw_events")
		cleanTable("events")
		cleanTable("aggregated_flows_by_dst")
		cleanTable("aggregated_flows_by_src")
	})

	insertRawEvent(
		"2020-10-11T07:00:00Z",
		"user1",
		"alexelcapo",
		"view",
	)
	insertRawEvent(
		"2020-10-11T07:00:00Z",
		"user2",
		"alexelcapo",
		"view",
	)
	insertRawEvent(
		"2020-10-11T07:00:00Z",
		"user3",
		"alexelcapo",
		"view",
	)
	insertRawEvent(
		"2020-10-11T07:00:00Z",
		"user4",
		"alexelcapo",
		"view",
	)
	insertRawEvent(
		"2020-10-11T07:00:00Z",
		"user5",
		"alexelcapo",
		"view",
	)
	insertRawEvent(
		"2020-10-11T08:00:00Z",
		"user1",
		"jujalag",
		"view",
	)
	insertRawEvent(
		"2020-10-11T08:00:00Z",
		"user2",
		"jujalag",
		"view",
	)
	insertRawEvent(
		"2020-10-11T08:00:00Z",
		"user3",
		"felipez",
		"view",
	)
	insertRawEvent(
		"2020-10-11T08:00:00Z",
		"user4",
		"jujalag",
		"view",
	)
	insertRawEvent(
		"2020-10-11T09:00:00Z",
		"user5",
		"felipez",
		"view",
	)
	insertRawEvent(
		"2020-10-11T09:00:00Z",
		"user6",
		"alexelcapo",
		"view",
	)
	insertRawEvent(
		"2020-10-11T11:00:00Z",
		"user6",
		"yuste",
		"view",
	)
	if err := ReconcileEvents(db, time.Time{}, 2*time.Hour); err != nil {
		t.Fatal(err)
	}

	got, err := UserFlowsBySrcHourly(
		db,
		"alexelcapo",
		parseTime("2020-10-11T08:00:00Z"),
		parseTime("2020-10-11T12:00:00Z"),
	)
	if err != nil {
		t.Fatal(err)
	}

	want := []*UserFlowSrc{
		{Ts: parseTime("2020-10-11T08:00:00Z"), Channel: "jujalag", Total: 3},
		{Ts: parseTime("2020-10-11T8:00:00Z"), Channel: "felipez", Total: 1},
		{Ts: parseTime("2020-10-11T9:00:00Z"), Channel: "felipez", Total: 1},
		{Ts: parseTime("2020-10-11T11:00:00Z"), Channel: "yuste", Total: 1},
	}
	if diff := deep.Equal(got, want); diff != nil {
		t.Fatal(diff)
	}
}

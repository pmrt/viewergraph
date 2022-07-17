package helix

import (
	"time"
)

// Twitch stream types
// See "Stream Online Event" https://dev.twitch.tv/docs/eventsub/eventsub-reference#events
const (
	StreamLive       string = "live"
	StreamPlaylist   string = "playlist"
	StreamWatchParty string = "watch_party"
	StreamPremiere   string = "premiere"
	StreamRerun      string = "rerun"
)

// Twitch Events
// See https://dev.twitch.tv/docs/eventsub/eventsub-reference#events

type BroadcasterDetails struct {
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
}

type EventStreamOnline struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	StartedAt time.Time `json:"stated_at"`
	BroadcasterDetails
}

type EventStreamOffline struct {
	BroadcasterDetails
}

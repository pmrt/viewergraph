package helix

import "database/sql"

// 1. Suscribirse a eventos
// 2. Registrar cbs, escuchar eventos (webhook) y ejecutar cbs
// 3. Gestionar credenciales (clientid/secret y token refresh)

type Helix struct {
	handleStreamOnline  func(evt *EventStreamOnline)
	handleStreamOffline func(evt *EventStreamOffline)

	handleRevocation func(evt *WebhookRevokePayload)
}

// OnStreamOnline sets the StreamOnline handler. The same event may be triggered
// more than once.
//
// https://dev.twitch.tv/docs/eventsub/eventsub-reference/#stream-online-event
func (hx *Helix) OnStreamOnline(cb func(evt *EventStreamOnline)) {
	hx.handleStreamOnline = cb
}

// OnStreamOffline sets the StreamOffline handler. The same event may be triggered
// more than once.
//
// https://dev.twitch.tv/docs/eventsub/eventsub-reference/#stream-offline-event
func (hx *Helix) OnStreamOffline(cb func(evt *EventStreamOffline)) {
	hx.handleStreamOffline = cb
}

func (hx *Helix) OnRevocation(cb func(evt *WebhookRevokePayload)) {
	hx.handleRevocation = cb
}

type Storage interface {
}

type StoragePostgres struct {
	db *sql.DB
}

func New() *Helix {
	return &Helix{}
}

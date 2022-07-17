package helix

import "database/sql"

// 1. Suscribirse a eventos
// 2. Registrar cbs, escuchar eventos (webhook) y ejecutar cbs
// 3. Gestionar credenciales (clientid/secret y token refresh)

type Helix struct {
	handleStreamOnline  func(evt *EventStreamOnline)
	handleStreamOffline func(evt *EventStreamOffline)
}

func (hx *Helix) OnStreamOnline(cb func(evt *EventStreamOnline)) {
	hx.handleStreamOnline = cb
}

func (hx *Helix) OnStreamOffline(cb func(evt *EventStreamOffline)) {
	hx.handleStreamOffline = cb
}

type Storage interface {
}

type StoragePostgres struct {
	db *sql.DB
}

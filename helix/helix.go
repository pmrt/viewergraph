package helix

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// 1. Suscribirse a eventos
// 2. Registrar cbs, escuchar eventos (webhook) y ejecutar cbs
// 3. Gestionar credenciales (clientid/secret y token refresh)

type Helix struct {
	clientID, secret         string
	APIUrl, EventSubEndpoint string

	c *http.Client

	handleStreamOnline  func(evt *EventStreamOnline)
	handleStreamOffline func(evt *EventStreamOffline)

	handleRevocation func(evt *WebhookRevokePayload)
}

const EstimatedSubscriptionJSONSize = 350

func (hx *Helix) CreateEventSubSubscription(sub *Subscription) error {
	b := struct {
		Type      string     `json:"type"`
		Version   string     `json:"version"`
		Condition *Condition `json:"condition"`
		Transport *Transport `json:"transport"`
	}{
		Type:      sub.Type,
		Version:   sub.Version,
		Condition: sub.Condition,
		Transport: sub.Transport,
	}

	buf := bytes.NewBuffer(make([]byte, 0, EstimatedSubscriptionJSONSize))
	if err := json.NewEncoder(buf).Encode(b); err != nil {
		return err
	}
	req, err := http.NewRequest(
		"POST",
		hx.APIUrl+hx.EventSubEndpoint+"/subscriptions",
		buf,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := hx.c.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("Expected 200 response, got" + fmt.Sprint(resp.StatusCode))
	}
	return nil
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

// Webhook handler for gofiber
func (hx *Helix) WebhookHandler(webhookSecret []byte) func(c *fiber.Ctx) error {
	h := &WebhookHandler{
		hx:     hx,
		secret: webhookSecret,
	}
	return h.handler
}

type Storage interface {
}

type StoragePostgres struct {
	db *sql.DB
}

func New(clientID, secret string) *Helix {
	return &Helix{
		clientID:         clientID,
		secret:           secret,
		APIUrl:           "https://api.twitch.tv/helix",
		EventSubEndpoint: "/eventsub",
	}
}

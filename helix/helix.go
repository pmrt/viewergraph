package helix

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/twitch"
)

// 1. Suscribirse a eventos
// 2. Registrar cbs, escuchar eventos (webhook) y ejecutar cbs
// 3. Gestionar credenciales (clientid/secret y token refresh)
type ClientCreds struct {
	ClientID, ClientSecret string
}

type Helix struct {
	ctx                      context.Context
	creds                    ClientCreds
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

// Exchange uses the client credentials to get a new http client with the
// corresponding token source, refreshing the token when needed. This http
// client injects the required Authorization header to the requests and will be
// used by the following requests.
//
// Must be used before using authenticated endpoints.
func (hx *Helix) Exchange() {
	o2 := &clientcredentials.Config{
		ClientID:     hx.creds.ClientID,
		ClientSecret: hx.creds.ClientSecret,
		TokenURL:     twitch.Endpoint.TokenURL,
	}
	hx.c = o2.Client(hx.ctx)
}

// NewWithoutExchange instantiates a new Helix client but without exchanging
// credentials for a token source. Useful for testing.
//
// Use New() if your helix client will be using authenticated endpoints.
func NewWithoutExchange(creds ClientCreds) *Helix {
	return &Helix{
		creds:            creds,
		ctx:              context.Background(),
		APIUrl:           "https://api.twitch.tv/helix",
		EventSubEndpoint: "/eventsub",
	}
}

func New(creds ClientCreds) *Helix {
	hx := NewWithoutExchange(creds)
	hx.Exchange()
	return hx
}

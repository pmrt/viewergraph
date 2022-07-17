package helix

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pmrt/viewergraph/utils"
)

var (
	WebhookEventHMACPrefix = []byte("sha256=")
)

// Twitch webhook event type
// See https://dev.twitch.tv/docs/eventsub/handling-webhook-events
const (
	WebhookEventNotification string = "notification"
	WebhookEventVerification string = "webhook_callback_verification"
	WebhookEventRevocation   string = "revocation"

	SubStreamOnline  string = "stream.online"
	SubStreamOffline string = "stream.offline"
)

// Twitch webhook headers
// https://dev.twitch.tv/docs/eventsub/handling-webhook-events#list-of-request-headers
const (
	WebhookHeaderID        = "Twitch-Eventsub-Message-Id"
	WebhookHeaderTimestamp = "Twitch-Eventsub-Message-Timestamp"
	WebhookHeaderSignature = "Twitch-Eventsub-Message-Signature"
	WebhookHeaderType      = "Twitch-Eventsub-Message-Type"
)

type WebhookEvent struct {
	ID        string
	Timestamp string
	Signature string
	Type      string
	Body      []byte
}

func (evt *WebhookEvent) Valid(secret []byte) bool {
	// Important note: DO NOT mutate id, sig and ts, they are meant to be read-only
	var (
		id   = utils.StringToByte(evt.ID)
		ts   = utils.StringToByte(evt.Timestamp)
		sig  = utils.StringToByte(evt.Signature)
		body = evt.Body
	)

	mac := hmac.New(sha256.New, secret)
	mac.Write(id)
	mac.Write(ts)
	mac.Write(body)
	hash := mac.Sum(nil)
	hexHash := make([]byte, 0, hex.EncodedLen(len(hash)+len(WebhookEventHMACPrefix)))
	hexHash = append(hexHash, WebhookEventHMACPrefix...)
	hexHash = append(hexHash, hash...)
	return hmac.Equal(sig, hexHash)
}

type WebhookHandler struct {
	secret []byte
	hx     *Helix
}

type WebhookNotificationResponse struct {
	Subscription struct {
		ID        string `json:"id"`
		Status    string `json:"status"`
		Type      string `json:"type"`
		Version   int    `json:"version"`
		Cost      int    `json:"cost"`
		Condition struct {
			BroadcasterUserID string `json:"broadcaster_user_id"`
		}
		Transport struct {
			Method   string `json:"method"`
			Callback string `json:"callback"`
		}
		CreatedAt time.Time `json:"created_at"`
	}
	Event struct {
	}
}

// {
//   "event": {
//     "user_id": "1337",
//     "user_login": "awesome_user",
//     "user_name": "Awesome_User",
//     "broadcaster_user_id":     "12826",
//     "broadcaster_user_login":  "twitch",
//     "broadcaster_user_name":   "Twitch"
//   }
// }

func (h *WebhookHandler) handler(c *fiber.Ctx) error {
	evt := &WebhookEvent{
		ID:        c.Get(WebhookHeaderID),
		Timestamp: c.Get(WebhookHeaderTimestamp),
		Signature: c.Get(WebhookHeaderSignature),
		Type:      c.Get(WebhookHeaderType),
		Body:      c.Body(),
	}
	if !evt.Valid(h.secret) {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid signature")
	}

	switch evt.Type {
	case WebhookEventNotification:
		var res *WebhookNotificationResponse
		if err := c.BodyParser(&res); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid notification body")
		}

		switch res.Subscription.Type {
		case SubStreamOnline:
			h.hx.handleStreamOnline()
		case SubStreamOffline:
			h.hx.handleStreamOffline()
		default:
			return fiber.NewError(fiber.StatusBadRequest, "Unknown notification subscription type")
		}

	case WebhookEventVerification:
	case WebhookEventRevocation:
	default:
		return fiber.NewError(fiber.StatusBadRequest, "Unknown Twitch-Eventsub-Message-Type header")
	}

	return nil
}

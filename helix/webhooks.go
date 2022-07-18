package helix

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pmrt/viewergraph/utils"
)

var (
	WebhookEventHMACPrefix       = []byte("sha256=")
	WebhookEventHMACPrefixLength = len(WebhookEventHMACPrefix)
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

type WebhookHeaders struct {
	ID        string
	Timestamp string
	Signature string
	Type      string
	Body      []byte
}

func (evt *WebhookHeaders) Valid(secret []byte) bool {
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
	l := len(hash)
	hexHash := make([]byte, hex.EncodedLen(l), hex.EncodedLen(l)+WebhookEventHMACPrefixLength)
	hex.Encode(hexHash, hash)
	hexHash = utils.Prepend(hexHash, WebhookEventHMACPrefix)
	return hmac.Equal(sig, hexHash)
}

type WebhookNotificationPayload struct {
	Subscription struct {
		ID        string `json:"id"`
		Status    string `json:"status"`
		Type      string `json:"type"`
		Version   string `json:"version"`
		Cost      int    `json:"cost"`
		Condition struct {
			BroadcasterUserID string `json:"broadcaster_user_id"`
		}
		Transport struct {
			Method   string `json:"method"`
			Callback string `json:"callback"`
		}
		CreatedAt time.Time `json:"created_at"`
	} `json:"subscription"`
	Event struct {
		ID                   string    `json:"id"`
		Type                 string    `json:"type"`
		StartedAt            time.Time `json:"started_at"`
		BroadcasterUserID    string    `json:"broadcaster_user_id"`
		BroadcasterUserLogin string    `json:"broadcaster_user_login"`
		BroadcasterUserName  string    `json:"broadcaster_user_name"`
	} `json:"event"`
}

type WebhookVerificationPayload struct {
	Challenge    string `json:"challenge"`
	Subscription struct {
		ID        string `json:"id"`
		Status    string `json:"status"`
		Type      string `json:"type"`
		Version   string `json:"version"`
		Cost      int    `json:"cost"`
		Condition struct {
			BroadcasterUserID string `json:"broadcaster_user_id"`
		}
		Transport struct {
			Method   string `json:"method"`
			Callback string `json:"callback"`
		}
		CreatedAt time.Time `json:"created_at"`
	} `json:"subscription"`
}

type WebhookHandler struct {
	secret []byte
	hx     *Helix
}

func (h *WebhookHandler) handler(c *fiber.Ctx) error {
	headers := &WebhookHeaders{
		ID:        c.Get(WebhookHeaderID),
		Timestamp: c.Get(WebhookHeaderTimestamp),
		Signature: c.Get(WebhookHeaderSignature),
		Type:      c.Get(WebhookHeaderType),
		Body:      c.Body(),
	}
	if !headers.Valid(h.secret) {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid signature")
	}

	// Note on replay attacks: twitch recommends ignoring events with ts>10min
	// and with previous IDs. But this requires storing the state of events IDs.
	// Instead we will allow channels to be tracked only once at the same time.
	switch headers.Type {
	case WebhookEventNotification:
		var resp *WebhookNotificationPayload
		if err := c.BodyParser(&resp); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid notification body")
		}

		switch resp.Subscription.Type {
		case SubStreamOnline:
			h.hx.handleStreamOnline(&EventStreamOnline{
				ID:        resp.Event.ID,
				Type:      resp.Event.Type,
				StartedAt: resp.Event.StartedAt,
				Broadcaster: &Broadcaster{
					ID:       resp.Event.BroadcasterUserID,
					Login:    resp.Event.BroadcasterUserLogin,
					Username: resp.Event.BroadcasterUserName,
				},
			})
		case SubStreamOffline:
			h.hx.handleStreamOffline(&EventStreamOffline{
				&Broadcaster{
					ID:       resp.Event.BroadcasterUserID,
					Login:    resp.Event.BroadcasterUserLogin,
					Username: resp.Event.BroadcasterUserName,
				},
			})
		default:
			return fiber.NewError(fiber.StatusBadRequest, "Unknown notification subscription type")
		}
	case WebhookEventVerification:
		var resp *WebhookVerificationPayload
		if err := c.BodyParser(&resp); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid notification body")
		}
		if resp.Challenge == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Empty challenge")
		}
		return c.SendString(resp.Challenge)
	case WebhookEventRevocation:
		// TODO - check revocation and re-sub if it is a tracked channel
	default:
		return fiber.NewError(fiber.StatusBadRequest, "Unknown Twitch-Eventsub-Message-Type header")
	}

	return nil
}

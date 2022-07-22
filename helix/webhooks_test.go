package helix

import (
	"bytes"
	"io/ioutil"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-test/deep"
	"github.com/gofiber/fiber/v2"
	"github.com/pmrt/viewergraph/config"
)

func TestWebhookHeadersValidation(t *testing.T) {
	t.Parallel()

	secret := []byte("zdsTKGJtGUiJyLMh5JRYCztpgppQh8Lo")

	tests := []struct {
		input *WebhookHeaders
		want  bool
	}{
		{
			input: &WebhookHeaders{
				ID:        "f1c2a387-161a-49f9-a165-0f21d7a4e1c4",
				Timestamp: "2019-11-16T10:11:12.123Z",
				Signature: "sha256=efff62e8394965726992ca425ac5aa9550b4e524e98b936b6bdddc2e86d53990",
				Type:      "notification",
				Body:      []byte("{body:1}"),
			},
			want: true,
		},
		{
			input: &WebhookHeaders{
				ID:        "f1c2a387-161a-49f9-a165-0f21d7a4e1c4",
				Timestamp: "2019-11-16T10:11:12.123Z",
				Signature: "sha256=efff62e8394965726992ca425ac5aa9550b4e524e98b936b6bdddc2e86d53990",
				Type:      "notification",
				Body:      []byte("{body:2}"),
				// Change:                ^
			},
			want: false,
		},
		{
			input: &WebhookHeaders{
				ID:        "f1c2a387-161a-49f9-a165-0f21d7a4e1c4",
				Timestamp: "2019-11-16T10:11:12.124Z",
				// Change:                        ^
				Signature: "sha256=efff62e8394965726992ca425ac5aa9550b4e524e98b936b6bdddc2e86d53990",
				Type:      "notification",
				Body:      []byte("{body:1}"),
			},
			want: false,
		},
		{
			input: &WebhookHeaders{
				ID: "f1c2a387-161a-49f9-a165-1f21d7a4e1c4",
				// Change:                   ^
				Timestamp: "2019-11-16T10:11:12.124Z",
				Signature: "sha256=efff62e8394965726992ca425ac5aa9550b4e524e98b936b6bdddc2e86d53990",
				Type:      "notification",
				Body:      []byte("{body:1}"),
			},
			want: false,
		},
	}

	for _, test := range tests {
		got := test.input.Valid(secret)
		if got != test.want {
			t.Fatalf("got %t, want %t", got, test.want)
		}
	}
}

var secret = []byte("thisisanososecretsecret")

func TestWebhookStreamOnline(t *testing.T) {
	t.Parallel()

	var onlineEvt *EventStreamOnline

	var body = []byte(`{
    "subscription": {
        "id": "f1c2a387-161a-49f9-a165-0f21d7a4e1c4",
        "type": "stream.online",
        "version": "1",
        "status": "enabled",
        "cost": 0,
        "condition": {
            "broadcaster_user_id": "1337"
        },
         "transport": {
            "method": "webhook",
            "callback": "https://example.com/webhooks/callback"
        },
        "created_at": "2019-11-16T10:11:12.123Z"
    },
    "event": {
        "id": "9001",
        "broadcaster_user_id": "1337",
        "broadcaster_user_login": "cool_user",
        "broadcaster_user_name": "Cool_User",
        "type": "live",
        "started_at": "2020-10-11T10:11:12.123Z"
    }
  }`)

	hx := NewWithoutExchange(ClientCreds{
		ClientID:     config.HelixClientID,
		ClientSecret: config.HelixSecret,
	})
	hx.OnStreamOnline(func(evt *EventStreamOnline) {
		onlineEvt = evt
	})

	app := fiber.New()
	app.Post("/webhook", hx.WebhookHandler(secret))

	req := httptest.NewRequest("POST", "http://localhost:7123/webhook", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(WebhookHeaderID, "f1c2a387-161a-49f9-a165-0f21d7a4e1c4")
	req.Header.Set(WebhookHeaderTimestamp, "2019-11-16T10:11:12.123Z")
	req.Header.Set(WebhookHeaderSignature, "sha256=135326f1ca01bb9ef7bb656053ce5a35e61a57ada77dc6705326c92d12c62060")
	req.Header.Set(WebhookHeaderType, WebhookEventNotification)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("\nexpected status code to be 200, got %d\nbody: %s", resp.StatusCode, b)
	}

	ts, err := time.Parse(time.RFC3339, "2020-10-11T10:11:12.123Z")
	if err != nil {
		t.Fatal(err)
	}

	got := onlineEvt
	want := &EventStreamOnline{
		ID:        "9001",
		Type:      "live",
		StartedAt: ts,
		Broadcaster: &Broadcaster{
			ID:       "1337",
			Username: "Cool_User",
			Login:    "cool_user",
		},
	}
	if diff := deep.Equal(got, want); diff != nil {
		t.Fatal(diff)
	}
}

func TestWebhookStreamOffline(t *testing.T) {
	t.Parallel()

	var onlineEvt *EventStreamOffline

	var body = []byte(`{
    "subscription": {
        "id": "f1c2a387-161a-49f9-a165-0f21d7a4e1c4",
        "type": "stream.offline",
        "version": "1",
        "status": "enabled",
        "cost": 0,
        "condition": {
            "broadcaster_user_id": "1337"
        },
        "created_at": "2019-11-16T10:11:12.123Z",
         "transport": {
            "method": "webhook",
            "callback": "https://example.com/webhooks/callback"
        }
    },
    "event": {
        "broadcaster_user_id": "1337",
        "broadcaster_user_login": "cool_user",
        "broadcaster_user_name": "Cool_User"
    }
  }`)

	hx := NewWithoutExchange(ClientCreds{
		ClientID:     config.HelixClientID,
		ClientSecret: config.HelixSecret,
	})
	hx.OnStreamOffline(func(evt *EventStreamOffline) {
		onlineEvt = evt
	})

	app := fiber.New()
	app.Post("/webhook", hx.WebhookHandler(secret))

	req := httptest.NewRequest("POST", "http://localhost:7123/webhook", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(WebhookHeaderID, "f1c2a387-161a-49f9-a165-0f21d7a4e1c4")
	req.Header.Set(WebhookHeaderTimestamp, "2019-11-16T10:11:12.123Z")
	req.Header.Set(WebhookHeaderSignature, "sha256=ce414455c20a25609bc0c276a052f461df1c11f14b90de15962131d5a715d827")
	req.Header.Set(WebhookHeaderType, WebhookEventNotification)

	resp, _ := app.Test(req)
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("\nexpected status code to be 200, got %d\nbody: %s", resp.StatusCode, b)
	}

	got := onlineEvt
	want := &EventStreamOffline{
		Broadcaster: &Broadcaster{
			ID:       "1337",
			Username: "Cool_User",
			Login:    "cool_user",
		},
	}
	if diff := deep.Equal(got, want); diff != nil {
		t.Fatal(diff)
	}
}

func TestWebhookVerification(t *testing.T) {
	var body = []byte(`{
    "challenge": "pogchamp-kappa-360noscope-vohiyo",
    "subscription": {
      "id": "f1c2a387-161a-49f9-a165-0f21d7a4e1c4",
      "status": "webhook_callback_verification_pending",
      "type": "channel.follow",
      "version": "1",
      "cost": 1,
      "condition": {
        "broadcaster_user_id": "12826"
      },
      "transport": {
        "method": "webhook",
        "callback": "https://example.com/webhooks/callback"
      },
      "created_at": "2019-11-16T10:11:12.123Z"
    }
  }`)

	hx := NewWithoutExchange(ClientCreds{
		ClientID:     config.HelixClientID,
		ClientSecret: config.HelixSecret,
	})

	app := fiber.New()
	app.Post("/webhook", hx.WebhookHandler(secret))

	req := httptest.NewRequest("POST", "http://localhost:7123/webhook", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(WebhookHeaderID, "f1c2a387-161a-49f9-a165-0f21d7a4e1c4")
	req.Header.Set(WebhookHeaderTimestamp, "2019-11-16T10:11:12.123Z")
	req.Header.Set(WebhookHeaderSignature, "sha256=876c54205d7c1ccb6966106190026ac2fcd6457a1d1010b6e7017b921a1fb4fd")
	req.Header.Set(WebhookHeaderType, WebhookEventVerification)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("\nexpected status code to be 200, got %d\nbody: %s", resp.StatusCode, b)
	}

	want := "pogchamp-kappa-360noscope-vohiyo"
	if string(b) != "pogchamp-kappa-360noscope-vohiyo" {
		t.Fatalf("expected body to be %s, got %s instead", want, b)
	}
}

func TestWebhookRevocation(t *testing.T) {
	var body = []byte(`{
    "subscription": {
      "id": "f1c2a387-161a-49f9-a165-0f21d7a4e1c4",
      "status": "authorization_revoked",
      "type": "channel.follow",
      "cost": 1,
      "version": "1",
      "condition": {
        "broadcaster_user_id": "12826"
      },
      "transport": {
        "method": "webhook",
        "callback": "https://example.com/webhooks/callback"
      },
      "created_at": "2019-11-16T10:11:12.123Z"
    }
  }`)

	var revokedEvt *WebhookRevokePayload
	hx := NewWithoutExchange(ClientCreds{
		ClientID:     config.HelixClientID,
		ClientSecret: config.HelixSecret,
	})
	hx.OnRevocation(func(evt *WebhookRevokePayload) {
		revokedEvt = evt
	})

	app := fiber.New()
	app.Post("/webhook", hx.WebhookHandler(secret))

	req := httptest.NewRequest("POST", "http://localhost:7123/webhook", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(WebhookHeaderID, "f1c2a387-161a-49f9-a165-0f21d7a4e1c4")
	req.Header.Set(WebhookHeaderTimestamp, "2019-11-16T10:11:12.123Z")
	req.Header.Set(WebhookHeaderSignature, "sha256=af10d7b0b3ac2708a168f6471b8e71fbfe8ede81b480f4f3c7d240e6faf56208")
	req.Header.Set(WebhookHeaderType, WebhookEventRevocation)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("\nexpected status code to be 200, got %d\nbody: %s", resp.StatusCode, b)
	}

	if revokedEvt.Subscription.Status != "authorization_revoked" {
		t.Fatalf(
			"expected subscription status to be authorization_revoked, got %s",
			revokedEvt.Subscription.Status,
		)
	}
}

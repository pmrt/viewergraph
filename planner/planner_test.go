package planner

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-test/deep"
	"github.com/pmrt/viewergraph/gen/vg/public/model"
	"github.com/pmrt/viewergraph/helix"
)

func TestPlannerFlush(t *testing.T) {
	tracked := []*model.TrackedChannels{
		{BroadcasterID: "1"},
	}
	got := make([]*helix.Subscription, 0, len(tracked)*2)
	fakeTwitchEventSubServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Log(err)
		}

		var sub *helix.Subscription
		if err := json.Unmarshal(b, &sub); err != nil {
			t.Log(err)
		}

		got = append(got, sub)
	}))

	p := FromChannels(&PlannerOpts{
		WebhookServerURL: "http://localhost",
		WebhookEndpoint:  "/webhook",
		WebhookSecret:    "fake-webhook-secret",
		WebhookPort:      "9530",
	}, tracked)
	p.hx = helix.NewWithoutExchange(helix.ClientCreds{
		ClientID:     "fake-id",
		ClientSecret: "fake-secret",
	})
	p.hx.APIUrl = fakeTwitchEventSubServer.URL

	if got, want := len(p.queue), 1; got != want {
		t.Fatalf("expected channel queue: got: %d, want: %d", got, want)
	}

	p.flush()

	want := []*helix.Subscription{
		{
			Type:    helix.SubStreamOnline,
			Version: "1",
			Condition: &helix.Condition{
				BroadcasterUserID: "1",
			},
			Transport: &helix.Transport{
				Method:   "webhook",
				Callback: "http://localhost/webhook",
				Secret:   "fake-webhook-secret",
			},
		},
		{
			Type:    helix.SubStreamOffline,
			Version: "1",
			Condition: &helix.Condition{
				BroadcasterUserID: "1",
			},
			Transport: &helix.Transport{
				Method:   "webhook",
				Callback: "http://localhost/webhook",
				Secret:   "fake-webhook-secret",
			},
		},
	}
	if diff := deep.Equal(want, got); diff != nil {
		t.Fatal(diff)
	}

	if p.queue != nil {
		t.Fatal("expected channel queue to be empty")
	}
}

func TestPlanner(t *testing.T) {
	sv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	}))

	hx := helix.NewWithoutExchange(helix.ClientCreds{
		ClientID:     "fake-id",
		ClientSecret: "fake-secret",
	})
	hx.APIUrl = sv.URL

	// p := FromChannels(tracked)

}

package helix

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pmrt/viewergraph/config"
)

func TestHelixCreateEventSubSubscription(t *testing.T) {
	const (
		broadcasterid = "1234"
		cb            = "http://localhost/webhook"
		secret        = "thisisanososecretsecret"
	)

	wantJson := `{"type":"stream.online","version":"1","condition":{"broadcaster_user_id":"1234"},"transport":{"method":"webhook","callback":"http://localhost/webhook","secret":"thisisanososecretsecret"}}` + string('\n')

	var body []byte
	sv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Log(err)
		}
		body = b
	}))
	defer sv.Close()
	hx := &Helix{
		clientID:         config.HelixClientID,
		secret:           config.HelixSecret,
		c:                sv.Client(),
		APIUrl:           sv.URL,
		EventSubEndpoint: "/eventsub",
	}
	err := hx.CreateEventSubSubscription(&Subscription{
		Type:    SubStreamOnline,
		Version: "1",
		Condition: &Condition{
			BroadcasterUserID: broadcasterid,
		},
		Transport: &Transport{
			Method:   "webhook",
			Callback: cb,
			Secret:   secret,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	got, want := string(body), wantJson
	if got != want {
		t.Fatalf("got:\n\n%s (%d)\nwant:\n\n%s (%d)", got, len(got), want, len(want))
	}
}

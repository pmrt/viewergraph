package helix

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-test/deep"
	"github.com/pmrt/viewergraph/config"
)

func TestHelixCredentials(t *testing.T) {
	cid, cs := os.Getenv("TEST_CLIENT_ID"), os.Getenv("TEST_CLIENT_SECRET")

	if cid == "" || cs == "" {
		t.Skip("WARNING: TEST_CLIENT_ID and TEST_CLIENT_SECRET environment variables needed for this test, skipping. Re-run test with required environment variables.")
	}

	hx := New(ClientCreds{
		ClientID:     cid,
		ClientSecret: cs,
	})

	if hx.c == nil {
		t.Fatal("client is empty")
	}

	endpoint := fmt.Sprintf("/users?login=%s", "alexelcapo")
	req, err := http.NewRequest("GET", hx.APIUrl+endpoint, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Client-Id", hx.creds.ClientID)

	resp, err := hx.c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 response, got %d", resp.StatusCode)
	}

	wantJSON := []byte(`{"data":[{"id":"36138196","login":"alexelcapo","display_name":"alexelcapo","type":"","broadcaster_type":"partner","description":"Nací en el 87 y me gusta jugar a videojuegos.","profile_image_url":"https://static-cdn.jtvnw.net/jtv_user_pictures/78528288-6216-4e21-872b-7f415b602a9a-profile_image-300x300.png","offline_image_url":"https://static-cdn.jtvnw.net/jtv_user_pictures/bf455aac-4ce9-4daa-94a0-c6c0a1b2500d-channel_offline_image-1920x1080.png","view_count":79789494,"created_at":"2012-09-12T21:24:26Z"}]}`)
	// Check some fields that we know will most likely never change
	var got, want struct {
		Data []struct {
			ID    string `json:"id"`
			Login string `json:"login"`
		} `json:"data"`
	}

	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(wantJSON, &want); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(got.Data[0], want.Data[0]); diff != nil {
		t.Fatal(diff)
	}

	if resp.Request.Header.Get("Authorization") == "" {
		t.Fatal("expected authorization request header to not be empty")
	}
}

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
		creds: ClientCreds{
			ClientID:     config.HelixClientID,
			ClientSecret: config.HelixSecret,
		},
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

package planner

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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

func createEventStreamOnline(bid, login string) *helix.EventStreamOnline {
	return &helix.EventStreamOnline{
		Broadcaster: &helix.Broadcaster{
			// once hashed this string will be 0 the sleep phase to align
			Login: login,
			ID:    bid,
		},
	}
}

func createEventStreamOffline(bid, login string) *helix.EventStreamOffline {
	return &helix.EventStreamOffline{
		Broadcaster: &helix.Broadcaster{
			// once hashed this string will be 0 the sleep phase to align
			Login: login,
			ID:    bid,
		},
	}
}

// List of string numbers that hashed with FNV32 and mod 60 will be equal to 0
// 234783 558010 293612 262375 92286 445402 544330 796494 31267
func TestPlannerStopsWithTimeout(t *testing.T) {
	before := func(ctx context.Context, bid string) {
		t.Fatal("worker should never be executed")
	}
	worker := func(ctx context.Context, bid string) {
		panic("worker should never be executed")
	}

	p := New(&PlannerOpts{
		TrackInterval:      10 * time.Second,
		TrackOnlineTimeout: 0,
		WorkerTimeout:      10 * time.Second,
		WorkerFunc:         worker,
		beforeWorkerTest:   before,
		SkipAlign:          true,
	})

	p.OnStreamOnline(createEventStreamOnline("234783", "user"))

	if _, ok := p.active.Get("234783"); ok {
		t.Fatal("expected executor to have no active cycle for bid 234783")
	}
}

func TestPlannerStopsWithEnd(t *testing.T) {
	var wg sync.WaitGroup

	var p *Planner
	var count1 uint32
	worker := func(ctx context.Context, bid string) {
		defer wg.Done()
		if bid == "234783" {
			atomic.AddUint32(&count1, 1)
		}
	}

	p = New(&PlannerOpts{
		TrackInterval:      10 * time.Second,
		TrackOnlineTimeout: 3600 * time.Second,
		WorkerTimeout:      10 * time.Second,
		WorkerFunc:         worker,
		SkipAlign:          true,
	})

	wg.Add(1)
	go p.OnStreamOnline(createEventStreamOnline("234783", "user"))
	wg.Wait()

	if count1 != 1 {
		t.Fatalf("count1: got:%d want:1", count1)
	}

	if _, ok := p.active.Get("234783"); !ok {
		t.Fatal("expected executor to have an active cycle for bid 234783")
	}

	wg.Add(1)
	go p.OnStreamOffline(createEventStreamOffline("234783", "user"))
	wg.Wait()

	if count1 != 2 {
		t.Fatalf("count1: got:%d want:2", count1)
	}

	if _, ok := p.active.Get("234783"); ok {
		t.Fatal("expected executor to have no active cycle for bid 234783")
	}
}

func TestPlannerStopsWithCtxPlanner(t *testing.T) {
	var wg sync.WaitGroup

	var p *Planner
	var count1, count2 uint32
	worker := func(ctx context.Context, bid string) {
		defer wg.Done()
		if bid == "234783" {
			atomic.AddUint32(&count1, 1)
		}
		if bid == "558010" {
			if p != nil {
				p.Stop()
			}
			if ctx.Err() == nil {
				// This will be executed if the context is not cancelled (ie. if
				// p.Stop() does not work)
				atomic.AddUint32(&count2, 1)
			}
		}
	}

	p = New(&PlannerOpts{
		TrackInterval:      10 * time.Second,
		TrackOnlineTimeout: 3600 * time.Second,
		WorkerTimeout:      10 * time.Second,
		WorkerFunc:         worker,
		SkipAlign:          true,
	})

	wg.Add(2)
	go p.OnStreamOnline(createEventStreamOnline("234783", "user"))
	go p.OnStreamOnline(createEventStreamOnline("558010", "user"))
	wg.Wait()

	if count1 != 1 {
		t.Fatalf("count1: got:%d want:1", count1)
	}
	if count2 != 0 {
		t.Fatalf("count2: got:%d want:0", count2)
	}
}

func TestPlannerPreventDuplicates(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2)

	var count1, count2 uint32
	worker := func(ctx context.Context, bid string) {
		if bid == "234783" {
			atomic.AddUint32(&count1, 1)
		}
		if bid == "558010" {
			atomic.AddUint32(&count2, 1)
		}
		wg.Done()
	}

	p := New(&PlannerOpts{
		TrackInterval:      10 * time.Second,
		TrackOnlineTimeout: 24 * time.Hour,
		WorkerTimeout:      5 * time.Minute,
		WorkerFunc:         worker,
		SkipAlign:          true,
	})

	bids := []string{"234783", "234783", "234783", "558010", "558010", "558010"}

	for _, bid := range bids {
		go func(bid string) {
			p.OnStreamOnline(createEventStreamOnline(bid, "user"))
		}(bid)
	}
	wg.Wait()
	p.Stop()

	if count2 != 1 {
		t.Fatalf("count1: got:%d want:1", count1)
	}
	if count2 != 1 {
		t.Fatalf("count2: got:%d want:1", count2)
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

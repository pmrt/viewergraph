package planner

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	cmap "github.com/pmrt/concurrent-map/v3"
	"github.com/pmrt/viewergraph/gen/vg/public/model"
	"github.com/pmrt/viewergraph/helix"
	l "github.com/rs/zerolog/log"
)

type PlannerOpts struct {
	Creds helix.ClientCreds

	WebhookServerURL string
	WebhookEndpoint  string
	WebhookSecret    string
	WebhookPort      string

	TrackIntervalMinutes time.Duration
	TrackOnlineTimeout   time.Duration

	WorkerFunc func(bid string)
}

type endSig chan struct{}

type Planner struct {
	ctx  context.Context
	opts *PlannerOpts
	hx   *helix.Helix
	sv   *fiber.App

	// queue of channels to be tracked
	queue []*model.TrackedChannels
	// active workers
	active cmap.ConcurrentMap[endSig]
}

func (p *Planner) Start() error {
	if p.hx == nil {
		p.hx = helix.New(p.opts.Creds)
	}

	l := l.With().
		Str("context", "planner").
		Logger()

	l.Info().Msg("initializing planner")

	l.Debug().Msg("-> setting up webhook handlers")
	p.hx.OnStreamOnline(p.OnStreamOnline)
	// p.hx.OnStreamOffline()

	p.sv.Post(
		p.opts.WebhookEndpoint,
		p.hx.WebhookHandler([]byte(p.opts.WebhookSecret)),
	)
	go func() {
		l.Debug().Msg("-> starting webhook server")
		p.sv.Hooks().OnListen(func() error {
			l.Debug().Msgf("-> -> webhook server listening on %s", p.opts.WebhookPort)
			return nil
		})
		// TODO - TLS
		if err := p.sv.Listen(":" + p.opts.WebhookPort); err != nil {
			l.Fatal().Err(err).Msg("")
		}
	}()

	p.flush()
	return nil
}

// TODO - test
// OnStreamOnline() is the heart of the planner. It is meant to be invoked by
// stream.online events from the EventSub (Webhook) Twitch API.
//
// Ensures only one online event per streamer is being run at the same time by
// using the BroadcasterID as key in a concurrent sharded hash map which
// contains a end channel to be closed upon stream.offline events.
//
// It aligns the cycle to a specific minute, so the worker is always run at
// roughly the same minute. The minute depends on the broadcasterID, the
// broadcasterID is hashed and balanced, distributed evenly across the minute
// range [0,59]. Depending on the hash, the same broadcasterID will run the task
// always at the same balanced minute. As a result, the workers will run at
// balanced minutes throughtout the 60 possible minutes.
//
// Once the cycle started there is three ways to stop this goroutine: (1) By a
// close signal in the general planner context which will stop all the active
// goroutines (2) By a close signal in the end channel, sent when other
// goroutine receives a stream.offline event with the same BroadcasterID. (3) By
// a timeout, this timeout is meant to close the channel if the stream has not
// received a stream.offline event after an abnormally long time.
func (p *Planner) OnStreamOnline(evt *helix.EventStreamOnline) {
	end := make(endSig, 1)

	if !p.active.SetIfAbsent(evt.Broadcaster.ID, end) {
		return
	}

	// generate a uniform minute based on broadcaster ID
	min := balancedKey(evt.Broadcaster.ID, 60)
	time.Sleep(untilMinute(int(min)))

	// If close signals were sent (ie. channels were closed) during our arbitrary
	// sleep phase (from 0 to 59 minutes), abort.
	//
	// This helps to mitigate most duplicated workers for cases where, for
	// example, we sleep for 59 minutes, the streamer ends broadcast within that
	// time span which invokes OnStreamOffline() where we run the worker one more
	// time before sending the end signal and then here we would run the worker
	// again before starting the cycle where we would detect that the channel is
	// closed.
	//
	// Note that there is still a possibility of duplicated workers if the
	// stream.offline event arrives at the same time we are running the worker. If
	// this becomes a problem in the future we could store the ts when the last
	// worker started in the concurrent hashmap
	select {
	case <-p.ctx.Done():
	case <-end:
		return
	default:
	}

	// start ticker (before running worker first time so it doesn't get delayed by
	// worker)
	ticker := time.NewTicker(p.opts.TrackIntervalMinutes)
	timeout := time.NewTimer(p.opts.TrackOnlineTimeout)
	defer ticker.Stop()
	// run worker once
	go p.opts.WorkerFunc(evt.Broadcaster.ID)

	// start cycle
	for {
		select {
		case <-p.ctx.Done():
		case <-end:
		case <-timeout.C:
			return
		case <-ticker.C:
			go p.opts.WorkerFunc(evt.Broadcaster.ID)
		}
	}
}

func (p *Planner) OnStreamOffline(evt *helix.EventStreamOffline) {
	// Run worker one more time before closing
	go p.opts.WorkerFunc(evt.Broadcaster.ID)
	if end, ok := p.active.Pop(evt.Broadcaster.ID); ok {
		close(end)
	}
}

func (p *Planner) flush() {
	if p.queue == nil {
		return
	}

	l := l.With().
		Str("context", "planner").
		Logger()

	l.Info().Msgf("flushing channel queue (%d)", len(p.queue))
	for _, ch := range p.queue {
		l.Debug().Msgf("-> req. subscription: %s (stream.online)", ch.BroadcasterID)
		if err := p.hx.CreateEventSubSubscription(&helix.Subscription{
			Type:    helix.SubStreamOnline,
			Version: "1",
			Condition: &helix.Condition{
				BroadcasterUserID: ch.BroadcasterID,
			},
			Transport: &helix.Transport{
				Method:   "webhook",
				Callback: p.opts.WebhookServerURL + p.opts.WebhookEndpoint,
				Secret:   p.opts.WebhookSecret,
			},
		}); err != nil {
			l.Error().
				Err(err).
				Str("bid", ch.BroadcasterID).
				Msg("error while subscribing to stream.online")
		}

		l.Debug().Msgf("-> req. subscription: %s (stream.offline)", ch.BroadcasterID)
		if err := p.hx.CreateEventSubSubscription(&helix.Subscription{
			Type:    helix.SubStreamOffline,
			Version: "1",
			Condition: &helix.Condition{
				BroadcasterUserID: ch.BroadcasterID,
			},
			Transport: &helix.Transport{
				Method:   "webhook",
				Callback: p.opts.WebhookServerURL + p.opts.WebhookEndpoint,
				Secret:   p.opts.WebhookSecret,
			},
		}); err != nil {
			l.Error().
				Err(err).
				Str("bid", ch.BroadcasterID).
				Msg("error while subscribing to stream.offline")
		}
	}

	p.queue = nil
}

func New(opts *PlannerOpts) *Planner {
	return &Planner{
		opts:   opts,
		ctx:    context.Background(),
		sv:     fiber.New(),
		active: cmap.NewWithConcurrencyLevel[endSig](32),
	}
}

func FromChannels(opts *PlannerOpts, tracked []*model.TrackedChannels) *Planner {
	p := New(opts)
	p.queue = tracked
	return p
}

package planner

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/pmrt/viewergraph/database"
	"github.com/pmrt/viewergraph/repo/clickhouse"
)

var ErrUnexpectedProp = errors.New("unexpected property")

var (
	OpenBracket  = json.Delim('[')
	CloseBracket = json.Delim(']')
	OpenBrace    = json.Delim('{')
	CloseBrace   = json.Delim('}')

	Closing = []json.Delim{
		CloseBrace, CloseBrace,
	}
)

// StreamBatcher parses the JSON object from the unofficial endpoint:
// tmi.twitch.tv/group/user/<user>/chatters and handles batching and inserting
// to the storage layer.
//
// Parsing, batching and inserting are performed in streaming mode from a
// reader. They are inserted as they are read every MaxQueueSize items, after
// the end of the slice or after the entire object is read if items length <
// MaxQueueSize. MaxQueueSize is also the maximum size of items in memory that
// StreamBatcher will use, it does not load the entire slice of items if the
// length > MaxQueueSize.
//
// Each StreamBatcher is 1:1 to each channel streaming. StreamBatcher is not
// thread safe.
type StreamBatcher struct {
	queue       []string
	queueCount  uint64
	ChatterSize uint64
	flushCount  uint64
	size        uint64

	FlushFunc func(sto database.Storage, queue []string, channel string) error

	MaxQueueSize uint64
	Channel      string
	sto          database.Storage
}

// Enqueue the given `usr` item.
//
// If queue is empty it will allocate a new slice with the smallest allocation
// size possible. If b.ChatterSize is not set before or equals to 0,
// b.MaxQueueSize will be used.
func (b *StreamBatcher) Enqueue(usr string) {
	if b.queue == nil {
		// Estimate the smallest possible allocation size
		b.size = minWithDefault(b.ChatterSize, b.MaxQueueSize, b.MaxQueueSize)
		left := b.ChatterSize - b.size*b.flushCount
		// Since we are working with unsigned ints, a negative result (e.g.: if
		// ChatterSize=0 and there are more values than MaxQueueSize will trigger
		// flushes and increase flushCount, so left < 0) will make the
		// number start again from the largest number it can store. We take
		// advantage of this behavior: b.size will always be smaller.
		b.size = minWithDefault(left, b.size, b.MaxQueueSize)
		b.queue = make([]string, 0, b.size)
	}
	b.queue = append(b.queue, usr)
	b.queueCount++

	if b.queueCount == b.size {
		b.Flush()
	}
}

// Flush the queue. Flush is an idempotent operation.
func (b *StreamBatcher) Flush() {
	if b.queueCount == 0 {
		return
	}
	b.FlushFunc(b.sto, b.queue, b.Channel)

	b.queue = nil
	b.queueCount = 0
	b.flushCount++
}

func (b *StreamBatcher) Batch(r io.Reader) error {
	dec := json.NewDecoder(r)
	tk, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := tk.(json.Delim); !ok || delim != '{' {
		return fmt.Errorf("expected JSON object at first token, got %s", tk)
	}

	for dec.More() {
		tk, err = dec.Token()
		if err != nil {
			return err
		}

		switch tk {
		case "chatter_count":
			// If provided, set ChatterSize. If we know the exact length we will
			// save allocation size and we won't need additional flushes. See
			// Enqueue().
			//
			// Note that for these optimizations we need 'chatter_count' to be read
			// before 'chatters' property and since we are reading from a stream we
			// can't control the order. If for any reason twitch decides to change the
			// order of the object and provides 'chatter_count' after 'chatters' this
			// would be rendered useless.
			if err := dec.Decode(&b.ChatterSize); err != nil {
				return err
			}
			continue
		case "_links":
			if err := skip(dec); err != nil {
				return err
			}
		case "chatters":
			for dec.More() {
				tk, err = dec.Token()
				if err != nil {
					return err
				}

				switch tk {
				case OpenBracket, OpenBrace, CloseBracket, CloseBrace:
				case "broadcaster":
					if err := skip(dec); err != nil {
						return err
					}
				case "vips", "moderators", "viewers", "staff", "admins", "global_mods":
					for {
						tk, err = dec.Token()
						if err != nil {
							return err
						}
						if tk == CloseBracket {
							break
						}
						if usr, ok := tk.(string); ok {
							b.Enqueue(usr)
						}
					}
				default:
					return fmt.Errorf("%w: '%s'", ErrUnexpectedProp, tk)
				}
			}
		default:
			return fmt.Errorf("%w: '%s'", ErrUnexpectedProp, tk)
		}
	}

	for _, delim := range Closing {
		tk, err := dec.Token()
		if err != nil {
			return err
		}

		if tk != delim {
			return fmt.Errorf("closing: expected %s at offset %d, got %s", delim, dec.InputOffset(), tk)
		}
	}

	// If ChatterSize wasn't provided we need an extra flush to ensure we don't
	// leave any items unflushed
	b.Flush()
	return nil
}

func flusher(sto database.Storage, queue []string, channel string) error {
	return clickhouse.InsertViewers(sto.Conn(), &clickhouse.Viewers{
		Ts:      time.Now(),
		Viewers: queue,
		Channel: channel,
	})
}

func NewStreamBatcher(sto database.Storage, channel string, batchSize uint64) *StreamBatcher {
	return &StreamBatcher{
		MaxQueueSize: batchSize,
		FlushFunc:    flusher,
		Channel:      channel,
		sto:          sto,
	}
}

// min takes two numbers and returns the minimum number.
func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

// minWithDefault is the same as min but if a is unset, ie. a=0, it will default
// to `def`.
func minWithDefault(a, b, def uint64) uint64 {
	if a == 0 {
		return def
	}
	return min(a, b)
}

// max takes two numbers and returns the maximum number.
func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

func skip(dec *json.Decoder) error {
	n := 0
	for {
		tk, err := dec.Token()
		if err != nil {
			return err
		}

		switch tk {
		case OpenBracket, OpenBrace:
			n++
		case CloseBracket, CloseBrace:
			n--
		}

		if n == 0 {
			return nil
		}
	}
}

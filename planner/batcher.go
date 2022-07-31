package planner

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

var ErrUnexpectedProp = errors.New("unexpected property")

type StreamBatcher struct {
	queue       []string
	queueCount  uint64
	ChatterSize uint64
	flushCount  uint64
	size        uint64

	FlushFunc func(queue []string)

	MaxQueueSize uint64
}

// TODO - qué pasa si count no es 0 pero recibo más elementos que count?

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
	b.FlushFunc(b.queue)

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
		return errors.New("expected JSON object at first token")
	}

	for dec.More() {
		tk, err = dec.Token()
		if err != nil {
			return err
		}
		key, ok := tk.(string)
		if !ok {
			return fmt.Errorf("expected string at offset %d", dec.InputOffset())
		}

		switch key {
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
			continue
		case "chatters":
			for dec.More() {
				tk, err = dec.Token()
				if err != nil {
					return err
				}
				key, ok = tk.(string)
				if !ok {
					return fmt.Errorf("expected string at offset %d", dec.InputOffset())
				}

				switch key {
				case "broadcaster":
					continue
				case "vips":
				case "moderators":
				case "viewers":
				case "staff":
				case "admins":
				case "global_mods":
					for dec.More() {
						var val string
						if err := dec.Decode(&val); err != nil {
							return err
						}
						b.Enqueue(val)
					}
				default:
					return fmt.Errorf("%w: '%s'", ErrUnexpectedProp, key)
				}
			}
		default:
			return fmt.Errorf("%w: '%s'", ErrUnexpectedProp, key)
		}
	}

	tk, err = dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := tk.(json.Delim); !ok || delim != '}' {
		return errors.New("expected JSON object at last token")
	}

	// If ChatterSize wasn't provided we need an extra flush to ensure we don't
	// leave any items unflushed
	b.Flush()
	return nil
}

func flusher(queue []string) {

}

func NewStreamBatcher() *StreamBatcher {
	return &StreamBatcher{
		MaxQueueSize: 100000,
		FlushFunc:    flusher,
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

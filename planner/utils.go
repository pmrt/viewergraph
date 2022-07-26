package planner

import (
	"time"
)

// balancedKey takes an arbitrary `key` and a `bucketsNum`, returning the
// corresponding bucket number for the key so it is evenly distributed across
// the number of buckets.
//
// Resulting bucket values are in the range [0, `bucketsNum`-1].
func balancedKey(key string, bucketsNum uint32) uint32 {
	return fnv32(key) % bucketsNum
}

func untilMinuteWithTime(t time.Time, min int) time.Duration {
	d := min - t.Minute()
	if d < 0 {
		d += 60
	}
	return time.Duration(d) * time.Minute
}

func untilMinute(min int) time.Duration {
	return untilMinuteWithTime(time.Now(), min)
}

func fnv32(key string) uint32 {
	var hash uint32 = 2166136261
	const prime32 uint32 = 16777619
	l := len(key)
	for i := 0; i < l; i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}

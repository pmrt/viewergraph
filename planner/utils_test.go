package planner

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/pmrt/viewergraph/utils"
)

func TestBalancedKeyDistribution(t *testing.T) {
	const nbuckets = 60

	inputs := make([]string, 10000)
	for i := range inputs {
		inputs[i] = strconv.Itoa(rand.Intn(1000000))
	}

	m := make(map[uint32]int, nbuckets)
	for _, in := range inputs {
		m[balancedKey(in, nbuckets)] += 1
	}

	r := make([]int, 0, nbuckets)
	for k := range m {
		r = append(r, m[k])
	}

	if len(m) != nbuckets {
		t.Fatal("expected length of hashmap to be equal to the number of buckets provided")
	}

	got := utils.CV(r, false)
	var wantThreshold float64 = 10
	if got > wantThreshold {
		t.Fatalf("expected keys to have a even distribution with a CV < 10%% (coefficient of variation). got cv=%f%%", got)
	}
}

func TestUntilMinute(t *testing.T) {
	T := func(str string) time.Time {
		t, err := time.Parse(time.RFC3339, str)
		if err != nil {
			panic(err)
		}
		return t
	}

	tests := []struct {
		t    time.Time
		min  int
		want time.Duration
	}{
		{t: T("2022-06-22T15:00:00Z"), min: 0, want: 0 * time.Minute},
		{t: T("2022-06-22T15:00:00Z"), min: 30, want: 30 * time.Minute},
		{t: T("2022-06-22T15:00:00Z"), min: 59, want: 59 * time.Minute},
		{t: T("2022-06-22T15:30:00Z"), min: 29, want: 59 * time.Minute},
		{t: T("2022-06-22T15:30:00Z"), min: 21, want: 51 * time.Minute},
		{t: T("2022-06-22T15:30:00Z"), min: 0, want: 30 * time.Minute},
		{t: T("2022-06-22T15:30:00Z"), min: 31, want: 1 * time.Minute},
		{t: T("2022-06-22T15:15:00Z"), min: 10, want: 55 * time.Minute},
		{t: T("2022-06-22T15:59:00Z"), min: 0, want: 1 * time.Minute},
		{t: T("2022-06-22T15:59:00Z"), min: 1, want: 2 * time.Minute},
		{t: T("2022-06-22T15:59:00Z"), min: 58, want: 59 * time.Minute},
		{t: T("2022-06-22T15:59:00Z"), min: 15, want: 16 * time.Minute},
	}

	for _, test := range tests {
		got, want := untilMinuteWithTime(test.t, test.min), test.want
		if got != want {
			t.Fatalf("T: %s, min: %d. Got %s, want %s", test.t, test.min, got, want)
		}
	}
}

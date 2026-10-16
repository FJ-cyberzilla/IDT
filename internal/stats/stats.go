package stats

import (
	"math"
	"sort"
	"time"
)

type Summary struct {
	Count  int
	Min    time.Duration
	Median time.Duration
	P90    time.Duration
	Max    time.Duration
	Jitter time.Duration
}

func Compute(samples []time.Duration) Summary {
	if len(samples) == 0 {
		return Summary{}
	}
	s := make([]time.Duration, len(samples))
	copy(s, samples)
	sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })

	sum := Summary{
		Count:  len(s),
		Min:    s[0],
		Max:    s[len(s)-1],
		Median: s[len(s)/2],
		P90:    percentile(s, 0.90),
	}

	var mean float64
	for _, d := range s {
		mean += float64(d)
	}
	mean /= float64(len(s))
	var variance float64
	for _, d := range s {
		diff := float64(d) - mean
		variance += diff * diff
	}
	variance /= float64(len(s))
	sum.Jitter = time.Duration(math.Sqrt(variance))
	return sum
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(math.Ceil(p*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

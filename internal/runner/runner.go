package runner

import (
	"context"
	"math/rand"
	"strings"
	"sync"
	"time"

	"idt/internal/config"
	"idt/internal/dialer"
	"idt/internal/stats"
)

const maxConcurrent = 3

type Result struct {
	Dialer                config.Dialer
	Summary               stats.Summary
	Samples               []time.Duration
	IPs                   [][]string
	Total                 int
	EOFErrorsCount        int
	EOFReconnectTriggered bool
}

var (
	eofMutex  sync.Mutex
	eofEvents []eofEvent
)

type eofEvent struct {
	host      string
	timestamp time.Time
}

func recordEOFAndCheckReconnect(host string, opts config.Options) bool {
	if !opts.EOFReconnectEnabled {
		return false
	}
	eofMutex.Lock()
	defer eofMutex.Unlock()

	now := time.Now()
	window := opts.EOFTimeWindow
	if window <= 0 {
		window = 20 * time.Second
	}
	threshold := opts.EOFTriggerThreshold
	if threshold <= 0 {
		threshold = 3
	}

	valid := make([]eofEvent, 0, len(eofEvents))
	for _, ev := range eofEvents {
		if now.Sub(ev.timestamp) <= window {
			valid = append(valid, ev)
		}
	}
	valid = append(valid, eofEvent{host: host, timestamp: now})
	eofEvents = valid

	hosts := map[string]bool{}
	for _, ev := range eofEvents {
		hosts[ev.host] = true
	}

	if len(hosts) >= threshold {
		eofEvents = nil
		return true
	}
	return false
}

func NewDialer(d config.Dialer, timeout time.Duration) dialer.Dialer {
	switch d.Protocol {
	case config.ProtoStandard:
		return dialer.Standard{Address: d.Address, Timeout: timeout}
	case config.ProtoDoT:
		return dialer.DoT{Address: d.Address, Timeout: timeout}
	case config.ProtoDoH:
		return dialer.DoH{URL: d.Address, Timeout: timeout}
	}
	return nil
}

func Dial(ctx context.Context, d config.Dialer, target string, timeout time.Duration) (dialer.Answer, error) {
	return NewDialer(d, timeout).Dial(ctx, target)
}

func Run(ctx context.Context, dialers []config.Dialer, opts config.Options) []Result {
	if opts.MaxDials > 0 {
		total := len(dialers) * opts.Dials
		if total > opts.MaxDials {
			opts.Dials = opts.MaxDials / len(dialers)
			if opts.Dials < 1 {
				opts.Dials = 1
			}
		}
	}

	results := make([]Result, len(dialers))
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for i, d := range dialers {
		// Context-aware semaphore acquisition prevents hanging on cancellation
		select {
		case <-ctx.Done():
			return results
		case sem <- struct{}{}:
		}

		wg.Add(1)
		go func(idx int, dl config.Dialer) {
			defer func() { <-sem }()
			defer wg.Done()
			results[idx] = runOne(ctx, dl, opts)
		}(i, d)
	}
	wg.Wait()
	return results
}

func runOne(ctx context.Context, d config.Dialer, opts config.Options) Result {
	r := runOneAttempt(ctx, d, opts)
	if r.Summary.Count == 0 {
		// Error handler resets and retries the test once upon stall or failure
		time.Sleep(250 * time.Millisecond)
		r = runOneAttempt(ctx, d, opts)
		if r.Summary.Count == 0 {
			// Persistent failure after reset & retry: states "not reachable this time"
		}
	}
	return r
}

func runOneAttempt(ctx context.Context, d config.Dialer, opts config.Options) Result {
	r := Result{Dialer: d, Total: opts.Dials}
	dl := NewDialer(d, opts.Timeout)
	if dl == nil {
		return r
	}

	// Goroutine-local RNG eliminates global mutex contention
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	hint := opts.Dials
	if hint < 0 {
		hint = 0
	}
	r.Samples = make([]time.Duration, 0, hint)
	r.IPs = make([][]string, 0, hint)

	for i := 0; i < opts.Dials; i++ {
		target := opts.Target
		if target == "" {
			target = config.SafeTargets[i%len(config.SafeTargets)]
		}

		ans, err := dl.Dial(ctx, target)
		if err != nil {
			errStr := strings.ToLower(err.Error())
			if strings.Contains(errStr, "429") {
				break
			}
			if strings.Contains(errStr, "eof") || strings.Contains(errStr, "connection reset by peer") || strings.Contains(errStr, "closed") {
				r.EOFErrorsCount++
				if recordEOFAndCheckReconnect(d.Address, opts) {
					r.EOFReconnectTriggered = true
					time.Sleep(100 * time.Millisecond)
				}
			}
			continue
		}
		r.Samples = append(r.Samples, ans.RTT)
		ips := make([]string, 0, len(ans.IPs))
		for _, ip := range ans.IPs {
			ips = append(ips, ip.String())
		}
		r.IPs = append(r.IPs, ips)

		if i < opts.Dials-1 {
			jmin := int(opts.JitterMin.Milliseconds())
			jmax := int(opts.JitterMax.Milliseconds())
			var jitterVal int
			if jmax > jmin {
				jitterVal = jmin + rng.Intn(jmax-jmin)
			} else {
				jitterVal = jmin
			}
			select {
			case <-time.After(time.Duration(jitterVal) * time.Millisecond):
			case <-ctx.Done():
				r.Summary = stats.Compute(r.Samples)
				return r
			}
		}
	}

	r.Summary = stats.Compute(r.Samples)
	return r
}

func SortByMedian(rs []Result) []Result {
	out := make([]Result, len(rs))
	copy(out, rs)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0; j-- {
			a, b := out[j-1], out[j]
			swap := false
			if a.Summary.Count == 0 && b.Summary.Count > 0 {
				swap = true
			} else if a.Summary.Count > 0 && b.Summary.Count > 0 &&
				b.Summary.Median < a.Summary.Median {
				swap = true
			}
			if !swap {
				break
			}
			out[j-1], out[j] = out[j], out[j-1]
		}
	}
	return out
}

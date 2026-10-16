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
	Dialer  config.Dialer
	Summary stats.Summary
	Samples []time.Duration
	IPs     [][]string
	Total   int
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
		wg.Add(1)
		go func(idx int, dl config.Dialer) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[idx] = runOne(ctx, dl, opts)
		}(i, d)
	}
	wg.Wait()
	return results
}

func runOne(ctx context.Context, d config.Dialer, opts config.Options) Result {
	r := Result{Dialer: d, Total: opts.Dials}
	dl := NewDialer(d, opts.Timeout)
	if dl == nil {
		return r
	}

	for i := 0; i < opts.Dials; i++ {
		target := opts.Target
		if target == "" {
			target = config.SafeTargets[i%len(config.SafeTargets)]
		}

		ans, err := dl.Dial(ctx, target)
		if err != nil {
			if strings.Contains(err.Error(), "429") {
				break
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
			var d int
			if jmax > jmin {
				d = jmin + rand.Intn(jmax-jmin)
			} else {
				d = jmin
			}
			select {
			case <-time.After(time.Duration(d) * time.Millisecond):
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

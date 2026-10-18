package vault

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"idt/internal/config"
	"idt/internal/dialer"
)

// LivePool manages the dynamic resolver pool, warm-up probing, oracle baselining, and picking.
type LivePool struct {
	mu              sync.RWMutex
	liveDialers     []config.Dialer
	fallbacks       []config.Dialer
	logger          Logger
	datalistDir     string
	probeTimeout    time.Duration
	target          string
	currentIndex    uint64
	lastRefresh     time.Time
	refreshInterval time.Duration
	stopChan        chan struct{}
}

// NewLivePool creates a new LivePool with hardcoded fallbacks and instrumentation logger.
func NewLivePool(logger Logger, datalistDir string) *LivePool {
	if logger == nil {
		logger = NoopLogger{}
	}
	fallbacks := []config.Dialer{
		{"Cloudflare-DoH", config.ProtoDoH, "https://cloudflare-dns.com/dns-query"},
		{"Google-Standard", config.ProtoStandard, "8.8.8.8:53"},
		{"Quad9-DoH", config.ProtoDoH, "https://dns.quad9.net/dns-query"},
	}

	return &LivePool{
		fallbacks:       fallbacks,
		logger:          logger,
		datalistDir:     datalistDir,
		probeTimeout:    2 * time.Second,
		target:          "example.com",
		refreshInterval: 15 * time.Minute,
		stopChan:        make(chan struct{}),
	}
}

// Warmup runs the warm-up pipeline: loads candidates, probes via UDP port 53, baselines against oracle, builds live pool.
func (p *LivePool) Warmup(ctx context.Context) error {
	candidates, err := LoadCandidates(p.datalistDir)
	if err != nil {
		p.logger.LogSnapshot(0, 0, 0, map[string]int{"load_error": 1})
		return fmt.Errorf("failed to load candidates: %w", err)
	}

	totalIn := len(candidates)
	dropCounts := make(map[string]int)
	var liveAfterProbe []config.Dialer

	// Baseline oracle: Cloudflare DoH
	oracle := dialer.DoH{
		URL:     "https://cloudflare-dns.com/dns-query",
		Timeout: p.probeTimeout,
	}
	expectedAnswer, oracleErr := oracle.Dial(ctx, p.target)
	hasOracle := oracleErr == nil

	for _, cand := range candidates {
		// Warm-up probe strictly pinned to plain UDP port 53
		stdDialer := dialer.Standard{
			Address: cand.Address,
			Timeout: p.probeTimeout,
		}

		start := time.Now()
		ans, dialErr := stdDialer.Dial(ctx, p.target)
		rtt := time.Since(start)

		if dialErr != nil {
			dropCounts["timeout_or_error"]++
			p.logger.LogWarmup(cand.Address, cand.Name, "error", rtt, dialErr.Error(), "timeout_or_error")
			continue
		}

		if len(ans.IPs) == 0 {
			dropCounts["empty_ips"]++
			p.logger.LogWarmup(cand.Address, cand.Name, "empty_ips", rtt, "", "empty_ips")
			continue
		}

		// Baseline comparison if oracle available
		if hasOracle && len(expectedAnswer.IPs) > 0 {
			match := false
			for _, expIP := range expectedAnswer.IPs {
				for _, gotIP := range ans.IPs {
					if expIP.Equal(gotIP) {
						match = true
						break
					}
				}
				if match {
					break
				}
			}
			if !match {
				dropCounts["baseline_mismatch"]++
				p.logger.LogWarmup(cand.Address, cand.Name, "mismatch", rtt, "", "baseline_mismatch")
				continue
			}
		}

		liveAfterProbe = append(liveAfterProbe, cand)
		p.logger.LogWarmup(cand.Address, cand.Name, "ok", rtt, "", "")
	}

	p.mu.Lock()
	p.liveDialers = liveAfterProbe
	p.lastRefresh = time.Now()
	p.mu.Unlock()

	liveCount := len(liveAfterProbe)
	p.logger.LogSnapshot(totalIn, liveCount, liveCount, dropCounts)

	if liveCount < 3 {
		log.Printf("⚠️ ALERT: Live resolver pool size (%d) is below safety threshold of 3!", liveCount)
	}

	return nil
}

// StartAutoRefresh starts a background goroutine refreshing the pool every 15 minutes.
func (p *LivePool) StartAutoRefresh(ctx context.Context) {
	ticker := time.NewTicker(p.refreshInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				_ = p.Warmup(ctx)
			case <-p.stopChan:
				ticker.Stop()
				return
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop stops the auto-refresh goroutine.
func (p *LivePool) Stop() {
	close(p.stopChan)
}

// Pick selects a resolver using round-robin over live dialers, falling back to fallbacks if live pool is empty.
func (p *LivePool) Pick() config.Dialer {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.liveDialers) > 0 {
		idx := atomic.AddUint64(&p.currentIndex, 1)
		return p.liveDialers[int(idx-1)%len(p.liveDialers)]
	}

	if len(p.fallbacks) > 0 {
		idx := atomic.AddUint64(&p.currentIndex, 1)
		return p.fallbacks[int(idx-1)%len(p.fallbacks)]
	}

	return config.Dialer{Name: "Cloudflare", Protocol: config.ProtoDoH, Address: "https://cloudflare-dns.com/dns-query"}
}

// Query performs a query using a picked live resolver, logging the test event.
func (p *LivePool) Query(ctx context.Context, target string) (dialer.Answer, string, error) {
	d := p.Pick()
	var (
		ans dialer.Answer
		err error
	)

	start := time.Now()
	switch d.Protocol {
	case config.ProtoDoH:
		doh := dialer.DoH{URL: d.Address, Timeout: 3 * time.Second}
		ans, err = doh.Dial(ctx, target)
	case config.ProtoStandard:
		std := dialer.Standard{Address: d.Address, Timeout: 3 * time.Second}
		ans, err = std.Dial(ctx, target)
	default:
		doh := dialer.DoH{URL: "https://cloudflare-dns.com/dns-query", Timeout: 3 * time.Second}
		ans, err = doh.Dial(ctx, target)
	}
	rtt := time.Since(start)

	outcome := "ok"
	var errStr string
	var ipStrs []string
	if err != nil {
		outcome = "error"
		errStr = err.Error()
	} else {
		for _, ip := range ans.IPs {
			ipStrs = append(ipStrs, ip.String())
		}
	}

	p.logger.LogTest(d.Name, target, "vault", outcome, ipStrs, nil, rtt, errStr)

	return ans, d.Name, err
}

// LiveCount returns the current number of live dialers in the pool.
func (p *LivePool) LiveCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.liveDialers)
}

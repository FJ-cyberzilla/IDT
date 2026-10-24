package singbox

import (
	"context"
	"net"
	"sync"
	"time"
)

// SafeguardConfig configures rate limiting and non-throttling routing bounds for low network environments.
type SafeguardConfig struct {
	MaxConcurrentDial int           `json:"max_concurrent_dial"`
	DialTimeout       time.Duration `json:"dial_timeout"`
	EnableThrottling  bool          `json:"enable_throttling"`
}

// DefaultSafeguardConfig provides safe defaults for low-network-advantage environments.
var DefaultSafeguardConfig = SafeguardConfig{
	MaxConcurrentDial: 5,
	DialTimeout:       3 * time.Second,
	EnableThrottling:  true,
}

// Safeguard provides a sing-box inspired routing and concurrency limiter to prevent
// system throttling or conflicts across Android, Termux, WSL2, and Windows.
type Safeguard struct {
	mu          sync.Mutex
	cfg         SafeguardConfig
	sem         chan struct{}
	dialer      net.Dialer
}

// NewSafeguard creates a new Safeguard instance.
func NewSafeguard(cfg SafeguardConfig) *Safeguard {
	if cfg.MaxConcurrentDial <= 0 {
		cfg.MaxConcurrentDial = DefaultSafeguardConfig.MaxConcurrentDial
	}
	if cfg.DialTimeout <= 0 {
		cfg.DialTimeout = DefaultSafeguardConfig.DialTimeout
	}

	return &Safeguard{
		cfg: cfg,
		sem: make(chan struct{}, cfg.MaxConcurrentDial),
		dialer: net.Dialer{
			Timeout: cfg.DialTimeout,
		},
	}
}

// DialContext safely executes a network dial with concurrency throttling and timeout bounds.
func (s *Safeguard) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if s.cfg.EnableThrottling {
		select {
		case s.sem <- struct{}{}:
			defer func() { <-s.sem }()
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return s.dialer.DialContext(ctx, network, address)
}

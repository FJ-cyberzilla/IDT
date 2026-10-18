package helper

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"runtime/debug"
	"sync"
	"time"
)

// SafeRun executes a function with panic recovery, converting any panic into a structured error
// and logging stack trace info to prevent silent crashes in network diagnostic routines.
func SafeRun(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			err = fmt.Errorf("panic recovered in network helper: %v\nstack:\n%s", r, stack)
		}
	}()
	return fn()
}

// SafeGo spawns a goroutine with panic recovery and context safety.
func SafeGo(ctx context.Context, fn func(context.Context)) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Log or handle background panic gracefully without crashing app
				fmt.Printf("[CRITICAL] background goroutine panic recovered: %v\n", r)
			}
		}()
		fn(ctx)
	}()
}

// Retry executes an operation with exponential backoff and randomized jitter to prevent thundering herds
// during automated network resilience testing.
func Retry(ctx context.Context, maxRetries int, initialDelay, maxDelay time.Duration, fn func(context.Context) error) error {
	var lastErr error
	delay := initialDelay

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("retry aborted by context: %w", err)
		}

		if err := fn(ctx); err == nil {
			return nil
		} else {
			lastErr = err
		}

		if attempt == maxRetries {
			break
		}

		// Calculate jittered exponential backoff
		// jitterFactor between 0.8 and 1.2
		jitterFactor := 0.8 + 0.4*rng.Float64()
		sleepDur := time.Duration(float64(delay) * jitterFactor)
		if sleepDur > maxDelay {
			sleepDur = maxDelay
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("retry interrupted: %w", ctx.Err())
		case <-time.After(sleepDur):
		}

		// Exponential step
		nextDelay := float64(delay) * 2.0
		if nextDelay > float64(maxDelay) {
			delay = maxDelay
		} else {
			delay = time.Duration(nextDelay)
		}
	}

	return fmt.Errorf("operation failed after %d retries, last error: %w", maxRetries, lastErr)
}

// GlobalSharedTransport returns an optimized HTTP transport with robust connection pooling,
// aggressive idle timeout pruning, and tuned socket parameters for high-frequency DNS-over-HTTPS (DoH).
var (
	sharedTransportOnce sync.Once
	globalTransport     *http.Transport
)

func GetSharedTransport() *http.Transport {
	sharedTransportOnce.Do(func() {
		globalTransport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          256,
			MaxIdleConnsPerHost:   64,
			MaxConnsPerHost:       128,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
	})
	return globalTransport
}

// IsPrivateIP checks whether a given IP address belongs to private/local ranges.
func IsPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	// Check loopback, link-local, private IPv4/IPv6
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	privCIDRs := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"100.64.0.0/10", // CGNAT
		"fc00::/7",      // Unique local IPv6
		"fe80::/10",     // Link-local IPv6
	}

	for _, cidr := range privCIDRs {
		_, subnet, err := net.ParseCIDR(cidr)
		if err == nil && subnet.Contains(ip) {
			return true
		}
	}
	return false
}

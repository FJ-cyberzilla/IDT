package helper

import (
	"context"
	"sync"
	"time"
)

// RateLimiter implements a token-bucket rate limiter to prevent upstream flagging and rate-limiting (429).
type RateLimiter struct {
	mu         sync.Mutex
	tokens     float64
	capacity   float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

// NewRateLimiter creates a rate limiter with a specified burst capacity and queries per second (QPS).
func NewRateLimiter(qps float64, burst int) *RateLimiter {
	if qps <= 0 {
		qps = 20.0
	}
	if burst <= 0 {
		burst = 30
	}
	return &RateLimiter{
		tokens:     float64(burst),
		capacity:   float64(burst),
		refillRate: qps,
		lastRefill: time.Now(),
	}
}

// Wait blocks until a token is available or context is cancelled.
func (rl *RateLimiter) Wait(ctx context.Context) error {
	for {
		rl.mu.Lock()
		now := time.Now()
		elapsed := now.Sub(rl.lastRefill).Seconds()
		rl.lastRefill = now
		rl.tokens += elapsed * rl.refillRate
		if rl.tokens > rl.capacity {
			rl.tokens = rl.capacity
		}

		if rl.tokens >= 1.0 {
			rl.tokens -= 1.0
			rl.mu.Unlock()
			return nil
		}

		needed := (1.0 - rl.tokens) / rl.refillRate
		rl.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(needed * float64(time.Second))):
		}
	}
}

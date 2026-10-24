package guardcat

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestGuardcatEndpointLifecycle(t *testing.T) {
	cfg := Config{
		FailureThreshold:    3,
		EOFTriggerThreshold: 3,
		ProbeInterval:       time.Second,
		MaxHeapBytes:        1024 * 1024 * 1024,
		EnableAutoGC:        true,
	}

	g := New(cfg)
	g.RegisterEndpoint("Cloudflare", "1.1.1.1:53", "udp")

	// Initially healthy
	if !g.IsHealthy("1.1.1.1:53", "udp") {
		t.Errorf("Expected endpoint to be healthy initially")
	}

	// Record failures (1 and 2 -> degraded)
	g.RecordFailure("1.1.1.1:53", "udp", errors.New("timeout"))
	g.RecordFailure("1.1.1.1:53", "udp", errors.New("timeout"))
	if !g.IsHealthy("1.1.1.1:53", "udp") {
		t.Errorf("Expected endpoint to be operational while degraded")
	}

	// 3rd failure -> unhealthy (circuit breaker tripped)
	g.RecordFailure("1.1.1.1:53", "udp", errors.New("timeout"))
	if g.IsHealthy("1.1.1.1:53", "udp") {
		t.Errorf("Expected endpoint to be unhealthy after 3 failures")
	}

	// Record success -> recovers
	g.RecordSuccess("1.1.1.1:53", "udp")
	if !g.IsHealthy("1.1.1.1:53", "udp") {
		t.Errorf("Expected endpoint to recover after success")
	}
}

func TestGuardcatEOFTriggerThreshold(t *testing.T) {
	cfg := Config{
		FailureThreshold:    5,
		EOFTriggerThreshold: 2, // Tripped after 2 EOF errors
	}
	g := New(cfg)

	g.RecordFailure("dns.example.com", "tls", errors.New("read tcp 127.0.0.1:1234: i/o timeout"))
	g.RecordFailure("dns.example.com", "tls", errors.New("remote error: tls: handshake failure")) // EOF-like or reset

	// Let's trigger EOF explicitly with "EOF"
	g.RecordFailure("dns.example.com", "tls", errors.New("io: EOF"))
	g.RecordFailure("dns.example.com", "tls", errors.New("connection reset by peer"))

	states := g.GetEndpointStates()
	st, ok := states["dns.example.com|tls"]
	if !ok {
		t.Fatalf("Expected endpoint state found")
	}

	if st.EOFCount < 2 {
		t.Errorf("Expected EOF count >= 2, got %d", st.EOFCount)
	}
	if st.Status != StatusUnhealthy {
		t.Errorf("Expected status UNHEALTHY due to EOF threshold, got %s", st.Status)
	}
}

func TestGuardcatMemoryAndReadiness(t *testing.T) {
	g := New(Config{MaxHeapBytes: 10 * 1024 * 1024 * 1024}) // 10 GB high watermark so it's OPTIMAL

	memRep := g.CheckMemory()
	if memRep.Status != "OPTIMAL" {
		t.Errorf("Expected memory status OPTIMAL, got %s", memRep.Status)
	}

	rec := g.RecommendAction()
	if rec == "" {
		t.Errorf("Expected non-empty recommendation string")
	}

	// Test readiness with a dummy or local address (may fail or succeed depending on environment, but test execution won't block)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_ = g.EnsureReadiness(ctx, "127.0.0.1:53")
}

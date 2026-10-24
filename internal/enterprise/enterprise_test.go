package enterprise

import (
	"context"
	"testing"
	"time"
)

func TestRunPerformanceMatrix(t *testing.T) {
	ctx := context.Background()
	results := RunPerformanceMatrix(ctx, DefaultLocations[:2], 3)

	expectedPairs := 2 * 2
	if len(results) != expectedPairs {
		t.Fatalf("expected %d matrix results, got %d", expectedPairs, len(results))
	}

	for _, r := range results {
		if r.Source.ID == "" || r.Destination.ID == "" {
			t.Errorf("invalid source or destination in matrix result")
		}
		if r.SLAStatus == "" {
			t.Errorf("missing SLA status")
		}
	}
}

func TestRunProtocolTests(t *testing.T) {
	ctx := context.Background()
	results := RunProtocolTests(ctx, DefaultProtocolTests[:2])

	if len(results) != 2 {
		t.Fatalf("expected 2 protocol test results, got %d", len(results))
	}

	for _, res := range results {
		if res.Score < 0 || res.Score > 100 {
			t.Errorf("invalid score %d for protocol %s", res.Score, res.Config.Protocol)
		}
		if res.Status == "" {
			t.Errorf("missing status for protocol %s", res.Config.Protocol)
		}
	}
}

func TestRunTrafficReplay(t *testing.T) {
	ctx := context.Background()
	cfg := ReplayConfig{
		SpeedMultiplier: 2.0,
		InjectJitter:    true,
		SimulateDrop:    false,
		BatchSize:       2,
	}

	res := RunTrafficReplay(ctx, SampleProductionTraffic[:3], cfg)

	if res.TotalCaptured != 3 {
		t.Errorf("expected 3 captured events, got %d", res.TotalCaptured)
	}
	if res.TotalReplayed != 3 {
		t.Errorf("expected 3 replayed events, got %d", res.TotalReplayed)
	}
	if len(res.LogSummary) != 3 {
		t.Errorf("expected 3 log summary entries, got %d", len(res.LogSummary))
	}
}

// Stress testing functions to ensure robustness under heavy load

func TestStressPerformanceMatrix(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Run full default locations matrix with high probe count under stress
	results := RunPerformanceMatrix(ctx, DefaultLocations, 50)
	expectedPairs := len(DefaultLocations) * len(DefaultLocations)
	if len(results) != expectedPairs {
		t.Fatalf("stress test failed: expected %d pairs, got %d", expectedPairs, len(results))
	}
}

func TestStressProtocolTests(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	heavyConfigs := make([]ProtocolTestConfig, 10)
	for i := 0; i < 10; i++ {
		heavyConfigs[i] = ProtocolTestConfig{
			Protocol:    ProtoHTTP,
			TargetName:  "Stress-Target-Gateway",
			Condition:   ConditionCongested,
			Iterations:  50,
		}
	}

	results := RunProtocolTests(ctx, heavyConfigs)
	if len(results) != 10 {
		t.Fatalf("stress protocol tests failed: expected 10 results, got %d", len(results))
	}
}

func TestStressTrafficReplay(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	largeTraffic := make([]TrafficEvent, 100)
	for i := 0; i < 100; i++ {
		largeTraffic[i] = TrafficEvent{
			ID:          "TR-STRESS",
			Source:      "Stress-Client",
			Endpoint:    "/api/stress",
			PayloadSize: 4096,
			OriginalRTT: 50 * time.Millisecond,
			HttpMethod:  "POST",
		}
	}

	cfg := ReplayConfig{
		SpeedMultiplier: 5.0,
		InjectJitter:    true,
		SimulateDrop:    true,
		BatchSize:       20,
	}

	res := RunTrafficReplay(ctx, largeTraffic, cfg)
	if res.TotalCaptured != 100 {
		t.Errorf("expected 100 captured, got %d", res.TotalCaptured)
	}
	if res.TotalReplayed+res.DroppedPackets != 100 {
		t.Errorf("packet accounting mismatch: replayed=%d, dropped=%d", res.TotalReplayed, res.DroppedPackets)
	}
}

func BenchmarkRunPerformanceMatrix(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		RunPerformanceMatrix(ctx, DefaultLocations[:3], 5)
	}
}

func BenchmarkRunProtocolTests(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		RunProtocolTests(ctx, DefaultProtocolTests)
	}
}

func BenchmarkRunTrafficReplay(b *testing.B) {
	ctx := context.Background()
	cfg := ReplayConfig{SpeedMultiplier: 1.0, InjectJitter: false, SimulateDrop: false}
	for i := 0; i < b.N; i++ {
		RunTrafficReplay(ctx, SampleProductionTraffic, cfg)
	}
}

func TestRunDNSCorruptionDetection(t *testing.T) {
	ctx := context.Background()
	res := RunDNSCorruptionDetection(ctx, DefaultCorruptionConfig)

	if res.TotalQueries != DefaultCorruptionConfig.Iterations {
		t.Errorf("expected %d queries, got %d", DefaultCorruptionConfig.Iterations, res.TotalQueries)
	}
	if res.Status == "" {
		t.Errorf("missing corruption detection status")
	}
}

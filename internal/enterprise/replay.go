package enterprise

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type TrafficEvent struct {
	ID          string        `json:"id"`
	Source      string        `json:"source"`
	Endpoint    string        `json:"endpoint"`
	PayloadSize int           `json:"payload_size_bytes"`
	OriginalRTT time.Duration `json:"original_rtt"`
	HttpMethod  string        `json:"http_method"`
}

type ReplayConfig struct {
	SpeedMultiplier float64 `json:"speed_multiplier"` // e.g., 1.0x, 2.0x, 5.0x
	InjectJitter    bool    `json:"inject_jitter"`
	SimulateDrop    bool    `json:"simulate_drop"`
	BatchSize       int     `json:"batch_size"`
}

type ReplayResult struct {
	TotalCaptured   int           `json:"total_captured"`
	TotalReplayed   int           `json:"total_replayed"`
	DroppedPackets  int           `json:"dropped_packets"`
	ReplayDuration  time.Duration `json:"replay_duration"`
	AvgReplayedRTT  time.Duration `json:"avg_replayed_rtt"`
	AnomalyDetected bool          `json:"anomaly_detected"`
	LogSummary      []string      `json:"log_summary"`
}

// SampleProductionTraffic returns a set of captured live production traffic records.
var SampleProductionTraffic = []TrafficEvent{
	{ID: "TR-101", Source: "Client-App-iOS", Endpoint: "/api/v1/auth/token", PayloadSize: 512, OriginalRTT: 18 * time.Millisecond, HttpMethod: "POST"},
	{ID: "TR-102", Source: "Web-Portal-SPA", Endpoint: "/api/v1/telemetry/stream", PayloadSize: 2048, OriginalRTT: 45 * time.Millisecond, HttpMethod: "PUT"},
	{ID: "TR-103", Source: "Microservice-B", Endpoint: "/api/v1/orders/batch", PayloadSize: 16384, OriginalRTT: 110 * time.Millisecond, HttpMethod: "POST"},
	{ID: "TR-104", Source: "Mobile-Android", Endpoint: "/api/v1/user/profile", PayloadSize: 1024, OriginalRTT: 28 * time.Millisecond, HttpMethod: "GET"},
	{ID: "TR-105", Source: "Partner-API-GW", Endpoint: "/webhook/enterprise/sync", PayloadSize: 8192, OriginalRTT: 85 * time.Millisecond, HttpMethod: "POST"},
}

// RunTrafficReplay executes a live production traffic capture and lab replay simulation.
func RunTrafficReplay(ctx context.Context, events []TrafficEvent, cfg ReplayConfig) ReplayResult {
	if len(events) == 0 {
		events = SampleProductionTraffic
	}
	if cfg.SpeedMultiplier <= 0 {
		cfg.SpeedMultiplier = 1.0
	}

	start := time.Now()
	var wg sync.WaitGroup
	var mu sync.Mutex

	replayedCount := 0
	droppedCount := 0
	var totalRTT time.Duration
	logs := make([]string, 0, len(events))
	anomaly := false

	sem := make(chan struct{}, 3)

	for i, ev := range events {
		// Context-aware semaphore acquisition prevents hanging on cancellation
		select {
		case <-ctx.Done():
			duration := time.Since(start)
			return ReplayResult{
				TotalCaptured:   len(events),
				TotalReplayed:   replayedCount,
				DroppedPackets:  droppedCount,
				ReplayDuration:  duration,
				AvgReplayedRTT:  0,
				AnomalyDetected: anomaly,
				LogSummary:      logs,
			}
		case sem <- struct{}{}:
		}

		wg.Add(1)
		go func(index int, e TrafficEvent) {
			defer func() { <-sem }()
			defer wg.Done()

			// Goroutine-local RNG eliminates global mutex contention
			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(index)))

			select {
			case <-ctx.Done():
				return
			default:
			}

			// Simulate processing / replay delay adjusted by speed multiplier
			simulatedRTT := time.Duration(float64(e.OriginalRTT) / cfg.SpeedMultiplier)
			if cfg.InjectJitter {
				jitter := time.Duration(rng.Intn(15)) * time.Millisecond
				simulatedRTT += jitter
			}

			isDropped := false
			if cfg.SimulateDrop && rng.Float64() < 0.15 {
				isDropped = true
			}

			mu.Lock()
			defer mu.Unlock()

			if isDropped {
				droppedCount++
				anomaly = true
				logs = append(logs, fmt.Sprintf("❌ DROPPED [Replay %s] %s %s (Payload: %dB)", e.ID, e.HttpMethod, e.Endpoint, e.PayloadSize))
			} else {
				replayedCount++
				totalRTT += simulatedRTT
				statusNote := "OK"
				if simulatedRTT > 90*time.Millisecond {
					anomaly = true
					statusNote = "HIGH LATENCY WARN"
				}
				logs = append(logs, fmt.Sprintf("✅ REPLAYED [Lab %s] %s %s -> RTT: %v [%s]", e.ID, e.HttpMethod, e.Endpoint, simulatedRTT, statusNote))
			}
		}(i, ev)
	}
	wg.Wait()

	duration := time.Since(start)
	var avgRTT time.Duration
	if replayedCount > 0 {
		avgRTT = totalRTT / time.Duration(replayedCount)
	}

	return ReplayResult{
		TotalCaptured:   len(events),
		TotalReplayed:   replayedCount,
		DroppedPackets:  droppedCount,
		ReplayDuration:  duration,
		AvgReplayedRTT:  avgRTT,
		AnomalyDetected: anomaly,
		LogSummary:      logs,
	}
}

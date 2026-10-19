package enterprise

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

// OfficeLocation represents a remote office or data center node.
type OfficeLocation struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Region      string        `json:"region"`
	BaseLatency time.Duration `json:"base_latency"`
	Jitter      time.Duration `json:"jitter"`
	PacketLoss  float64       `json:"packet_loss"` // percentage 0.0 - 100.0
	BandwidthMB float64       `json:"bandwidth_mb"`
}

// MatrixResult holds performance metrics between two locations.
type MatrixResult struct {
	Source      OfficeLocation        `json:"source"`
	Destination OfficeLocation        `json:"destination"`
	Latency     time.Duration         `json:"latency"`
	Jitter      time.Duration         `json:"jitter"`
	Loss        float64               `json:"loss"`
	Throughput  float64               `json:"throughput_mbps"`
	SLAStatus   string                `json:"sla_status"` // "OPTIMAL", "WARNING", "VIOLATED"
	Samples     []time.Duration       `json:"samples"`
}

// DefaultLocations provides standard enterprise remote offices and data centers.
var DefaultLocations = []OfficeLocation{
	{ID: "nyc-dc", Name: "New York Data Center", Region: "US-East", BaseLatency: 12 * time.Millisecond, Jitter: 2 * time.Millisecond, PacketLoss: 0.1, BandwidthMB: 1000.0},
	{ID: "sfo-dc", Name: "San Francisco Data Center", Region: "US-West", BaseLatency: 45 * time.Millisecond, Jitter: 5 * time.Millisecond, PacketLoss: 0.3, BandwidthMB: 1000.0},
	{ID: "lon-office", Name: "London Regional Office", Region: "EU-West", BaseLatency: 85 * time.Millisecond, Jitter: 8 * time.Millisecond, PacketLoss: 0.5, BandwidthMB: 500.0},
	{ID: "fra-dc", Name: "Frankfurt Hub DC", Region: "EU-Central", BaseLatency: 95 * time.Millisecond, Jitter: 7 * time.Millisecond, PacketLoss: 0.2, BandwidthMB: 1000.0},
	{ID: "tok-office", Name: "Tokyo Branch Office", Region: "AP-East", BaseLatency: 140 * time.Millisecond, Jitter: 12 * time.Millisecond, PacketLoss: 0.8, BandwidthMB: 300.0},
	{ID: "syd-office", Name: "Sydney Regional Hub", Region: "AP-South", BaseLatency: 180 * time.Millisecond, Jitter: 15 * time.Millisecond, PacketLoss: 1.0, BandwidthMB: 300.0},
}

// RunPerformanceMatrix simulates and measures performance across all office-to-datacenter pairs concurrently.
func RunPerformanceMatrix(ctx context.Context, locations []OfficeLocation, probesPerPair int) []MatrixResult {
	if len(locations) == 0 {
		locations = DefaultLocations
	}
	if probesPerPair <= 0 {
		probesPerPair = 5
	}

	totalPairs := len(locations) * len(locations)
	results := make([]MatrixResult, totalPairs)
	
	var wg sync.WaitGroup
	sem := make(chan struct{}, 4) // max 4 concurrent pairs

	idx := 0
	for i, src := range locations {
		for j, dst := range locations {
			if i == j {
				// Self loop
				results[idx] = MatrixResult{
					Source:      src,
					Destination: dst,
					Latency:     1 * time.Millisecond,
					Jitter:      0,
					Loss:        0.0,
					Throughput:  src.BandwidthMB,
					SLAStatus:   "OPTIMAL",
					Samples:     []time.Duration{1 * time.Millisecond},
				}
				idx++
				continue
			}

			sem <- struct{}{}
			wg.Add(1)
			go func(index int, s, d OfficeLocation) {
				defer func() { <-sem }()
				defer wg.Done()

				samples := make([]time.Duration, 0, probesPerPair)
				var totalLat time.Duration
				lostCount := 0

				base := (s.BaseLatency + d.BaseLatency) / 2
				jittMax := int((s.Jitter + d.Jitter).Milliseconds())
				if jittMax < 1 {
					jittMax = 1
				}
				lossRate := (s.PacketLoss + d.PacketLoss) / 2

				for p := 0; p < probesPerPair; p++ {
					select {
					case <-ctx.Done():
						return
					default:
					}

					// Simulate packet loss
					if rand.Float64()*100.0 < lossRate {
						lostCount++
						continue
					}

					jitterOffset := time.Duration(rand.Intn(jittMax)) * time.Millisecond
					sample := base + jitterOffset
					samples = append(samples, sample)
					totalLat += sample
				}

				actualLoss := (float64(lostCount) / float64(probesPerPair)) * 100.0
				var avgLat time.Duration
				if len(samples) > 0 {
					avgLat = totalLat / time.Duration(len(samples))
				} else {
					avgLat = base * 2
				}

				sla := "OPTIMAL"
				if avgLat > 120*time.Millisecond || actualLoss > 0.5 {
					sla = "WARNING"
				}
				if avgLat > 200*time.Millisecond || actualLoss > 2.0 {
					sla = "VIOLATED"
				}

				minBw := s.BandwidthMB
				if d.BandwidthMB < minBw {
					minBw = d.BandwidthMB
				}
				throughput := minBw * (1.0 - (actualLoss / 100.0))

				results[index] = MatrixResult{
					Source:      s,
					Destination: d,
					Latency:     avgLat,
					Jitter:      time.Duration(jittMax) * time.Millisecond,
					Loss:        actualLoss,
					Throughput:  throughput,
					SLAStatus:   sla,
					Samples:     samples,
				}
			}(idx, src, dst)
			idx++
		}
	}
	wg.Wait()
	return results[:idx]
}

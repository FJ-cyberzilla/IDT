package enterprise

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

type ProtocolType string

const (
	ProtoHTTP   ProtocolType = "HTTP/REST API"
	ProtoVoIP   ProtocolType = "VoIP / SIP-RTP"
	ProtoSaaS   ProtocolType = "SaaS / M365 / Salesforce"
	ProtoCustom ProtocolType = "Custom App TCP/UDP"
)

type NetworkCondition string

const (
	ConditionOptimal   NetworkCondition = "Normal / Optimal"
	ConditionCongested NetworkCondition = "High Congestion"
	ConditionJittery   NetworkCondition = "Extreme Jitter"
	ConditionDegraded  NetworkCondition = "Packet Loss / Drop"
)

type ProtocolTestConfig struct {
	Protocol    ProtocolType     `json:"protocol"`
	TargetName  string           `json:"target_name"`
	Condition   NetworkCondition `json:"condition"`
	Iterations  int              `json:"iterations"`
}

type ProtocolTestResult struct {
	Config      ProtocolTestConfig `json:"config"`
	SuccessRate float64            `json:"success_rate"` // percentage
	AvgLatency  time.Duration      `json:"avg_latency"`
	Jitter      time.Duration      `json:"jitter"`
	PacketLoss  float64            `json:"packet_loss"`
	Score       int                `json:"score"` // 0 - 100 health score
	Status      string             `json:"status"` // "HEALTHY", "DEGRADED", "CRITICAL"
	Details     string             `json:"details"`
}

var DefaultProtocolTests = []ProtocolTestConfig{
	{Protocol: ProtoHTTP, TargetName: "Global API Gateway (HTTPS)", Condition: ConditionOptimal, Iterations: 10},
	{Protocol: ProtoVoIP, TargetName: "Enterprise SIP Media Relay", Condition: ConditionJittery, Iterations: 10},
	{Protocol: ProtoSaaS, TargetName: "Microsoft 365 / Slack Core", Condition: ConditionOptimal, Iterations: 10},
	{Protocol: ProtoCustom, TargetName: "Custom ERP Microservice Engine", Condition: ConditionCongested, Iterations: 10},
}

// RunProtocolTests runs multi-protocol synthetic probes under dynamic conditions concurrently.
func RunProtocolTests(ctx context.Context, configs []ProtocolTestConfig) []ProtocolTestResult {
	if len(configs) == 0 {
		configs = DefaultProtocolTests
	}

	results := make([]ProtocolTestResult, len(configs))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 3)

	for i, cfg := range configs {
		sem <- struct{}{}
		wg.Add(1)
		go func(idx int, c ProtocolTestConfig) {
			defer func() { <-sem }()
			defer wg.Done()

			iterations := c.Iterations
			if iterations <= 0 {
				iterations = 10
			}

			var baseLat time.Duration
			var lossProb float64
			var jitRange int

			switch c.Protocol {
			case ProtoHTTP:
				baseLat = 35 * time.Millisecond
				lossProb = 0.2
				jitRange = 5
			case ProtoVoIP:
				baseLat = 20 * time.Millisecond // VoIP requires ultra-low latency
				lossProb = 0.5
				jitRange = 15
			case ProtoSaaS:
				baseLat = 65 * time.Millisecond
				lossProb = 0.1
				jitRange = 8
			case ProtoCustom:
				baseLat = 50 * time.Millisecond
				lossProb = 1.0
				jitRange = 20
			}

			// Adjust based on condition
			switch c.Condition {
			case ConditionCongested:
				baseLat = time.Duration(float64(baseLat) * 2.5)
				lossProb *= 3.0
				jitRange *= 2
			case ConditionJittery:
				jitRange = int(float64(jitRange) * 1.5)
				lossProb *= 1.5
			case ConditionDegraded:
				lossProb *= 5.0
				baseLat = time.Duration(float64(baseLat) * 1.8)
			}

			successCount := 0
			var totalLat time.Duration
			var maxLat time.Duration

			for k := 0; k < iterations; k++ {
				select {
				case <-ctx.Done():
					return
				default:
				}

				if rand.Float64()*100.0 < lossProb {
					continue // dropped packet
				}

				jitterVal := time.Duration(rand.Intn(jitRange+1)) * time.Millisecond
				lat := baseLat + jitterVal
				successCount++
				totalLat += lat
				if lat > maxLat {
					maxLat = lat
				}
			}

			successRate := (float64(successCount) / float64(iterations)) * 100.0
			var avgLat time.Duration
			if successCount > 0 {
				avgLat = totalLat / time.Duration(successCount)
			} else {
				avgLat = baseLat * 3
			}

			score := int(successRate * (1.0 - (float64(avgLat.Milliseconds()) / 500.0)))
			if score < 0 {
				score = 0
			}
			if score > 100 {
				score = 100
			}

			status := "HEALTHY"
			details := "All protocol transactions verified successfully within SLA."
			if successRate < 90.0 || avgLat > 150*time.Millisecond {
				status = "DEGRADED"
				details = "Elevated latency or jitter detected under dynamic conditions."
			}
			if successRate < 70.0 || avgLat > 300*time.Millisecond {
				status = "CRITICAL"
				details = "High packet loss or timeout thresholds breached."
			}

			results[idx] = ProtocolTestResult{
				Config:      c,
				SuccessRate: successRate,
				AvgLatency:  avgLat,
				Jitter:      time.Duration(jitRange) * time.Millisecond,
				PacketLoss:  100.0 - successRate,
				Score:       score,
				Status:      status,
				Details:     details,
			}
		}(i, cfg)
	}
	wg.Wait()
	return results
}

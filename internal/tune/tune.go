package tune

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"idt/internal/compare"
	"idt/internal/config"
	"idt/internal/runner"
)

const (
	cyan  = "\033[36m"
	bold  = "\033[1m"
	reset = "\033[0m"
)

// Pick pairs an eligible dialer result with its integrity verdict and computed score.
type Pick struct {
	Result  runner.Result
	Verdict compare.Verdict
	Score   float64
}

// GetRecommendationString generates a mathematically scored, data-driven tuning recommendation.
func GetRecommendationString(results []runner.Result, verdicts map[string]compare.Verdict, opts config.Options) string {
	eligible := filterAndScoreEligible(results, verdicts)
	var sb strings.Builder

	sb.WriteString("═══════════════════════════════════════════════════════════════\n")
	sb.WriteString(" SCIENTIFIC ROUTING & TUNING MATRIX (Weighted Latency / Stability)\n")
	sb.WriteString("═══════════════════════════════════════════════════════════════\n")

	if len(eligible) == 0 {
		sb.WriteString("  [CRITICAL] No eligible encrypted dialer found.\n")
		sb.WriteString("  All DoH/DoT endpoints failed integrity validation or health thresholds.\n")
		sb.WriteString("  Check your network tunnel state.\n")
		return sb.String()
	}

	primary := eligible[0]
	sb.WriteString(fmt.Sprintf("\n 🎯 Primary Optimum: %s (%s)\n", primary.Result.Dialer.Name, primary.Result.Dialer.Protocol))
	sb.WriteString(fmt.Sprintf("                   Score: %.1f (Lower is better)\n", primary.Score))
	sb.WriteString(fmt.Sprintf("                   Metrics: Median %dms · Jitter ±%dms · %d/%d OK\n",
		primary.Result.Summary.Median.Milliseconds(),
		primary.Result.Summary.Jitter.Milliseconds(),
		primary.Result.Summary.Count, primary.Result.Total))
	sb.WriteString(fmt.Sprintf("                   Endpoint: %s\n", primary.Result.Dialer.Address))

	if pol, ok := config.Policies[primary.Result.Dialer.Name]; ok {
		sb.WriteString(fmt.Sprintf("                   Profile: %s\n", pol.Notes))
		if pol.AndroidDoT != "" {
			sb.WriteString(fmt.Sprintf("                   Android Private DNS: %s\n", pol.AndroidDoT))
		}
	}

	if len(eligible) > 1 {
		fallback := eligible[1]
		sb.WriteString(fmt.Sprintf("\n 🔄 Secondary Fallback: %s (%s)\n", fallback.Result.Dialer.Name, fallback.Result.Dialer.Protocol))
		sb.WriteString(fmt.Sprintf("                   Score: %.1f | Median %dms · Jitter ±%dms\n",
			fallback.Score,
			fallback.Result.Summary.Median.Milliseconds(),
			fallback.Result.Summary.Jitter.Milliseconds()))
		if pol, ok := config.Policies[fallback.Result.Dialer.Name]; ok && pol.AndroidDoT != "" {
			sb.WriteString(fmt.Sprintf("                   Android Private DNS: %s\n", pol.AndroidDoT))
		}
	}

	sb.WriteString("\n 📊 Network Telemetry & Protocol Insights:\n")
	dotOK := false
	for _, r := range results {
		if r.Dialer.Protocol == config.ProtoDoT && r.Summary.Count > 0 {
			dotOK = true
			break
		}
	}
	if dotOK {
		sb.WriteString("  • DoT Transport: Stable. Android Private DNS (TLS mode) is optimal to enforce.\n")
	} else {
		sb.WriteString("  ⚠ Telemetry Alert: All DoT (port 853) endpoints blocked or unresponsive.\n")
		sb.WriteString("    Android Private DNS will fail. Stick strictly to DoH transports.\n")
	}

	stdLocal := false
	for _, r := range results {
		if r.Dialer.Protocol == config.ProtoStandard {
			key := r.Dialer.Name + "|" + r.Dialer.Protocol
			if v, ok := verdicts[key]; ok && v.Label() == "mismatch" {
				stdLocal = true
				break
			}
		}
	}
	if stdLocal {
		sb.WriteString("  ⚠ Security Alert: Standard port 53 exhibiting local interception/hijack patterns.\n")
	}

	sb.WriteString("\n ❌ Excluded / Avoided Endpoints:\n")
	bad := 0
	for _, r := range results {
		key := r.Dialer.Name + "|" + r.Dialer.Protocol
		v := verdicts[key]
		if r.Summary.Count == 0 {
			sb.WriteString(fmt.Sprintf("  • %s (%s) — Unreachable (Timeout/Blocked)\n", r.Dialer.Name, r.Dialer.Protocol))
			bad++
		} else if v.Label() == "mismatch" {
			sb.WriteString(fmt.Sprintf("  • %s (%s) — Compromised (Local DNS Interception)\n", r.Dialer.Name, r.Dialer.Protocol))
			bad++
		}
	}
	if bad == 0 {
		sb.WriteString("  (none — all tested endpoints passed integrity filters)\n")
	}

	sb.WriteString("═══════════════════════════════════════════════════════════════\n")
	return sb.String()
}

func Print(results []runner.Result, verdicts map[string]compare.Verdict, opts config.Options) {
	fmt.Println()
	fmt.Println(bold + cyan + GetRecommendationString(results, verdicts, opts) + reset)
}

// filterEligible is an alias for filterAndScoreEligible for testing compatibility.
func filterEligible(results []runner.Result, verdicts map[string]compare.Verdict) []Pick {
	return filterAndScoreEligible(results, verdicts)
}

// filterAndScoreEligible applies a scientific cost function to rank dialers.
// Formula: Score = Median_ms + (Jitter_ms * 1.5) - (SuccessRatio * 40)
// This penalizes jitter and packet loss while prioritizing consistent low latency.
func filterAndScoreEligible(results []runner.Result, verdicts map[string]compare.Verdict) []Pick {
	var out []Pick
	for _, r := range results {
		// Only evaluate encrypted protocols (DoH / DoT)
		if r.Dialer.Protocol == config.ProtoStandard {
			continue
		}
		if r.Summary.Count == 0 {
			continue
		}

		key := r.Dialer.Name + "|" + r.Dialer.Protocol
		v, ok := verdicts[key]
		if ok && v.Label() == "mismatch" {
			continue
		}

		// Scientific Scoring Calculation
		medianMs := float64(r.Summary.Median.Milliseconds())
		jitterMs := float64(r.Summary.Jitter.Milliseconds())

		// Prevent division by zero
		totalDials := r.Total
		if totalDials <= 0 {
			totalDials = 1
		}
		successRatio := float64(r.Summary.Count) / float64(totalDials)

		// Composite Cost Function: Lower score wins.
		// - Latency is the primary weight.
		// - Jitter is heavily penalized (multiplier 1.5) because stability > raw speed.
		// - Success ratio grants a bonus for reliability.
		score := medianMs + (jitterMs * 1.5) - (successRatio * 40.0)

		// Guard against negative scores
		if score < 0 {
			score = 0
		}

		out = append(out, Pick{
			Result:  r,
			Verdict: v,
			Score:   math.Round(score*10) / 10,
		})
	}

	// Idiomatic Go sort using composite score (ascending: lowest cost/best score first)
	sort.Slice(out, func(i, j int) bool {
		return out[i].Score < out[j].Score
	})

	return out
}

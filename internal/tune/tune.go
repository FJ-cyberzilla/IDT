package tune

import (
	"fmt"
	"strings"

	"idt/internal/compare"
	"idt/internal/config"
	"idt/internal/runner"
)

const (
	green  = "\033[32m"
	yellow = "\033[33m"
	red    = "\033[31m"
	cyan   = "\033[36m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	reset  = "\033[0m"
)

// Pick is one eligible dialer paired with its answer comparison.
type Pick struct {
	Result  runner.Result
	Verdict compare.Verdict
}

// GetRecommendationString generates the tuning recommendation text.
func GetRecommendationString(results []runner.Result, verdicts map[string]compare.Verdict, opts config.Options) string {
	eligible := filterEligible(results, verdicts)
	var sb strings.Builder

	sb.WriteString("═══════════════════════════════════════════════════════════════\n")
	sb.WriteString(" TUNING RECOMMENDATION for this device / current network\n")
	sb.WriteString("═══════════════════════════════════════════════════════════════\n")

	if len(eligible) == 0 {
		sb.WriteString("  No eligible encrypted dialer found.\n")
		sb.WriteString("  All DoH/DoT endpoints failed or answered locally.\n")
		sb.WriteString("  Check your network — this is unusual even on locked-down Wi-Fi.\n")
		return sb.String()
	}

	primary := eligible[0]
	sb.WriteString(fmt.Sprintf("\n Primary dialer:   %s (%s)\n", primary.Result.Dialer.Name, primary.Result.Dialer.Protocol))
	sb.WriteString(fmt.Sprintf("                   Median %dms · %d/%d dials OK · clean answers\n",
		primary.Result.Summary.Median.Milliseconds(),
		primary.Result.Summary.Count, primary.Result.Total))
	sb.WriteString(fmt.Sprintf("                   Endpoint: %s\n", primary.Result.Dialer.Address))

	if pol, ok := config.Policies[primary.Result.Dialer.Name]; ok {
		sb.WriteString(fmt.Sprintf("                   %s\n", pol.Notes))
		if pol.AndroidDoT != "" {
			sb.WriteString(fmt.Sprintf("                   Android Private DNS: %s\n", pol.AndroidDoT))
		}
	}

	if len(eligible) > 1 {
		fallback := eligible[1]
		sb.WriteString(fmt.Sprintf("\n Fallback dialer:  %s (%s)\n", fallback.Result.Dialer.Name, fallback.Result.Dialer.Protocol))
		sb.WriteString(fmt.Sprintf("                   Median %dms · %d/%d dials OK\n",
			fallback.Result.Summary.Median.Milliseconds(),
			fallback.Result.Summary.Count, fallback.Result.Total))
		if pol, ok := config.Policies[fallback.Result.Dialer.Name]; ok && pol.AndroidDoT != "" {
			sb.WriteString(fmt.Sprintf("                   Android Private DNS: %s\n", pol.AndroidDoT))
		}
	}

	sb.WriteString("\n Network notes & Settings warnings:\n")
	dotOK := false
	for _, r := range results {
		if r.Dialer.Protocol == config.ProtoDoT && r.Summary.Count > 0 {
			dotOK = true
			break
		}
	}
	if dotOK {
		sb.WriteString("  • DoT works → Android Private DNS is safe and optimal to enable.\n")
	} else {
		sb.WriteString("  ⚠ WARNING (Settings/Network Mismatch): DoT port 853 is blocked or restricted.\n")
		sb.WriteString("    Android Private DNS will NOT work here. Use DoH via a browser/app instead.\n")
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
		sb.WriteString("  ⚠ WARNING (Local Interception): Standard port 53 is answering locally (VPN/ISP).\n")
		sb.WriteString("    Standard dialer responses differ from expected upstream target.\n")
	}

	sb.WriteString("\n Avoid on this network:\n")
	bad := 0
	for _, r := range results {
		key := r.Dialer.Name + "|" + r.Dialer.Protocol
		v := verdicts[key]
		if r.Summary.Count == 0 {
			sb.WriteString(fmt.Sprintf("  • %s (%s) — unreachable\n", r.Dialer.Name, r.Dialer.Protocol))
			bad++
		} else if v.Label() == "mismatch" {
			sb.WriteString(fmt.Sprintf("  • %s (%s) — answering locally\n", r.Dialer.Name, r.Dialer.Protocol))
			bad++
		}
	}
	if bad == 0 {
		sb.WriteString("  (none)\n")
	}

	sb.WriteString("═══════════════════════════════════════════════════════════════\n")
	return sb.String()
}

// Print analyses results and prints a tuning recommendation.
// Only encrypted dialers (DoH/DoT) are eligible for the primary slot.
func Print(results []runner.Result, verdicts map[string]compare.Verdict, opts config.Options) {
	fmt.Println()
	fmt.Println(bold + cyan + GetRecommendationString(results, verdicts, opts) + reset)
}

// filterEligible narrows results to encrypted, fully successful,
// non-hijacked dialers, then sorts by median RTT ascending.
func filterEligible(results []runner.Result, verdicts map[string]compare.Verdict) []Pick {
	var out []Pick
	for _, r := range results {
		if r.Dialer.Protocol == config.ProtoStandard {
			continue
		}
		if r.Summary.Count == 0 {
			continue
		}
		if r.Summary.Count < r.Total {
			continue
		}
		key := r.Dialer.Name + "|" + r.Dialer.Protocol
		v, ok := verdicts[key]
		if ok && v.Label() == "mismatch" {
			continue
		}
		out = append(out, Pick{Result: r, Verdict: v})
	}

	for i := 1; i < len(out); i++ {
		for j := i; j > 0; j-- {
			if out[j].Result.Summary.Median < out[j-1].Result.Summary.Median {
				out[j], out[j-1] = out[j-1], out[j]
			} else {
				break
			}
		}
	}
	return out
}

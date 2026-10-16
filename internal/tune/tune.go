package tune

import (
	"fmt"

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

// Print analyses results and prints a tuning recommendation.
// Only encrypted dialers (DoH/DoT) are eligible for the primary slot.
func Print(results []runner.Result, verdicts map[string]compare.Verdict, opts config.Options) {
	eligible := filterEligible(results, verdicts)

	fmt.Println()
	fmt.Println(bold + cyan + "═══════════════════════════════════════════════════════════════" + reset)
	fmt.Printf("%s%s TUNING RECOMMENDATION for this device / current network%s\n",
		bold, cyan, reset)
	fmt.Println(bold + cyan + "═══════════════════════════════════════════════════════════════" + reset)

	if len(eligible) == 0 {
		fmt.Println(red + "  No eligible encrypted dialer found." + reset)
		fmt.Println("  All DoH/DoT endpoints failed or answered locally.")
		fmt.Println("  Check your network — this is unusual even on locked-down Wi-Fi.")
		return
	}

	primary := eligible[0]

	fmt.Printf("\n %sPrimary dialer:%s   %s%s%s (%s)\n",
		bold, reset, green, primary.Result.Dialer.Name, reset, primary.Result.Dialer.Protocol)
	fmt.Printf("                   Median %s%dms%s · %d/%d dials OK · clean answers\n",
		green, primary.Result.Summary.Median.Milliseconds(), reset,
		primary.Result.Summary.Count, primary.Result.Total)
	fmt.Printf("                   Endpoint: %s\n", primary.Result.Dialer.Address)

	if pol, ok := config.Policies[primary.Result.Dialer.Name]; ok {
		fmt.Printf("                   %s%s%s\n", dim, pol.Notes, reset)
		if pol.AndroidDoT != "" {
			fmt.Printf("                   Android Private DNS: %s%s%s\n",
				bold, pol.AndroidDoT, reset)
		}
	}

	if len(eligible) > 1 {
		fallback := eligible[1]
		fmt.Printf("\n %sFallback dialer:%s  %s (%s)\n",
			bold, reset, fallback.Result.Dialer.Name, fallback.Result.Dialer.Protocol)
		fmt.Printf("                   Median %dms · %d/%d dials OK\n",
			fallback.Result.Summary.Median.Milliseconds(),
			fallback.Result.Summary.Count, fallback.Result.Total)
		if pol, ok := config.Policies[fallback.Result.Dialer.Name]; ok && pol.AndroidDoT != "" {
			fmt.Printf("                   Android Private DNS: %s\n", pol.AndroidDoT)
		}
	}

	fmt.Printf("\n %sNetwork notes:%s\n", bold, reset)

	dotOK := false
	for _, r := range results {
		if r.Dialer.Protocol == config.ProtoDoT && r.Summary.Count > 0 {
			dotOK = true
			break
		}
	}
	if dotOK {
		fmt.Println("  • DoT works → Android Private DNS is safe to enable.")
	} else {
		fmt.Println("  • DoT blocked → Android Private DNS will NOT work here.")
		fmt.Println("    Use DoH via a browser or DNS app instead.")
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
		fmt.Println("  • Port 53 answers locally (VPN/ISP).")
		fmt.Println("    Standard dialers are not reaching their named targets.")
	}

	fmt.Printf("\n %sAvoid on this network:%s\n", bold, reset)
	bad := 0
	for _, r := range results {
		key := r.Dialer.Name + "|" + r.Dialer.Protocol
		v := verdicts[key]
		if r.Summary.Count == 0 {
			fmt.Printf("  • %s (%s) — unreachable\n", r.Dialer.Name, r.Dialer.Protocol)
			bad++
		} else if v.Label() == "mismatch" {
			fmt.Printf("  • %s (%s) — answering locally\n", r.Dialer.Name, r.Dialer.Protocol)
			bad++
		}
	}
	if bad == 0 {
		fmt.Println("  (none)")
	}

	fmt.Println(bold + cyan + "═══════════════════════════════════════════════════════════════" + reset)
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

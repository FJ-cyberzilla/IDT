package check

import (
	"context"
	"fmt"
	"time"

	"idt/internal/config"
	"idt/internal/runner"
)

// Dialers is the minimal health-check set — 5 total.
var Dialers = []config.Dialer{
	{"Cloudflare", config.ProtoDoH, "https://cloudflare-dns.com/dns-query"},
	{"Cloudflare", config.ProtoDoT, "one.one.one.one:853"},
	{"Cloudflare", config.ProtoStandard, "1.1.1.1:53"},
	{"Google", config.ProtoDoH, "https://dns.google/dns-query"},
	{"Quad9", config.ProtoDoH, "https://dns.quad9.net/dns-query"},
}

// Run performs a minimal safe health check and prints a plain-English verdict.
// Total: 5 dials, one iteration each, jittered, neutral target.
func Run(ctx context.Context) {
	opts := config.Options{
		Target:    "example.com",
		Dials:     1,
		Timeout:   3 * time.Second,
		JitterMin: 1000 * time.Millisecond,
		JitterMax: 2000 * time.Millisecond,
		MaxDials:  10,
	}

	fmt.Println("IDT — Internet Dialer Test")
	fmt.Println("Health check — example.com — 5 dials — sequential")
	fmt.Println()

	results := runner.Run(ctx, Dialers, opts)

	var (
		dohOK, dotOK, anySuccess bool
		stdLocal                 bool
	)

	for _, r := range results {
		status := "❌ no answer"
		latency := "-"
		if r.Summary.Count > 0 {
			anySuccess = true
			status = "✅ ok"
			latency = fmt.Sprintf("%d ms", r.Summary.Median.Milliseconds())
			switch r.Dialer.Protocol {
			case config.ProtoDoH:
				dohOK = true
			case config.ProtoDoT:
				dotOK = true
			case config.ProtoStandard:
				if r.Summary.Median < 3*time.Millisecond {
					stdLocal = true
				}
			}
		}
		fmt.Printf("  %-12s %-9s %-9s  %s\n",
			r.Dialer.Name, r.Dialer.Protocol, latency, status)
	}

	fmt.Println()
	fmt.Println("VERDICT:")
	switch {
	case !anySuccess:
		fmt.Println("  ❌ No dialer reachable. Name resolution is broken.")
		fmt.Println("     Check: VPN state, airplane mode, Wi-Fi/cellular.")
	case !dohOK:
		fmt.Println("  ❌ Encrypted dialers (DoH) are unreachable.")
		fmt.Println("     Check: firewall blocking 443, captive portal, VPN failure.")
	case !dotOK:
		fmt.Println("  ⚠  DoH works, but DoT (port 853) is blocked.")
		fmt.Println("     Android Private DNS will NOT work here.")
	default:
		fmt.Println("  ✅ Encrypted dialers (DoH + DoT) are reachable.")
	}
	if stdLocal {
		fmt.Println("  ⚠  Standard port 53 is answering locally (VPN/ISP).")
		fmt.Println("     The dialer you named is not the dialer you reached.")
	}
	if dohOK {
		fmt.Println()
		fmt.Println("  → If a site is not loading but encrypted dialers work,")
		fmt.Println("    the problem is NOT name resolution.")
	}
}

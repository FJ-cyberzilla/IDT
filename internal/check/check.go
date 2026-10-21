package check

import (
        "context"
        "fmt"

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

// Run performs a comprehensive health check aligned with options and dialers.
func Run(ctx context.Context) {
        opts, dialers, err := config.Load()
        if err != nil {
                opts = config.DefaultOptions()
                dialers = config.DefaultDialers()
        }

        fmt.Println("IDT — Internet Dialer Test (Fully Aligned Matrix)")
        fmt.Printf("Health check — target: %s — %d dialers configured\n", opts.Target, len(dialers))
        fmt.Printf("Options — Dials: %d, Timeout: %v, Jitter: %v–%v\n\n", 
                opts.Dials, opts.Timeout, opts.JitterMin, opts.JitterMax)

        results := runner.Run(ctx, dialers, opts)

        var (
                dohSuccessCount int
                dotSuccessCount int
                anySuccess      bool
                stdAnswering    bool
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
                                dohSuccessCount++
                        case config.ProtoDoT:
                                dotSuccessCount++
                        case config.ProtoStandard:
                                stdAnswering = true
                        }
                }
                fmt.Printf("  %-15s %-9s %-9s  %s\n",
                        r.Dialer.Name, r.Dialer.Protocol, latency, status)
        }

        fmt.Println()
        fmt.Println("VERDICT:")
        switch {
        case !anySuccess:
                fmt.Println("  ❌ All dialers unreachable. Name resolution is completely broken.")
                fmt.Println("     Check: VPN state, airplane mode, or physical network link.")
        case dohSuccessCount == 0:
                fmt.Println("  ❌ All DoH endpoints failed. Encrypted HTTPS DNS is blocked.")
                fmt.Println("     Check: Firewall blocking port 443, captive portal, or proxy failure.")
        case dotSuccessCount == 0:
                fmt.Println("  ⚠  DoH works, but all DoT (port 853) endpoints failed.")
                fmt.Println("     Android Private DNS (TLS mode) may struggle on this network.")
        default:
                fmt.Printf("  ✅ Enterprise matrix verified (DoH successes: %d, DoT successes: %d).\n", 
                        dohSuccessCount, dotSuccessCount)
        }

        if stdAnswering {
                fmt.Println("  ⚠  Standard port 53 (UDP) is responding on some endpoints.")
                fmt.Println("     Ensure your routing matches your privacy requirements.")
        }

        if dohSuccessCount > 0 {
                fmt.Println()
                fmt.Println("  → Encrypted transport channels are operational.")
                fmt.Println("    System is clear for secure proxy and API dispatch.")
        }
}

package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"

	"idt/internal/check"
	"idt/internal/compare"
	"idt/internal/config"
	"idt/internal/output"
	"idt/internal/runner"
	"idt/internal/tune"
)

func main() {
	opts, dialers, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config error:", err)
		os.Exit(1)
	}

	ctx := context.Background()

	if opts.Check {
		check.Run(ctx)
		return
	}

	if opts.Prefer != "" {
		dialers = config.FilterByPreference(dialers, opts.Prefer)
		if len(dialers) == 0 {
			fmt.Fprintln(os.Stderr, "no dialers match preference:", opts.Prefer)
			os.Exit(1)
		}
	}

	if opts.Plan {
		total := len(dialers) * opts.Dials
		if opts.MaxDials > 0 && total > opts.MaxDials {
			opts.Dials = opts.MaxDials / len(dialers)
			if opts.Dials < 1 {
				opts.Dials = 1
			}
			total = len(dialers) * opts.Dials
		}
		fmt.Println("PLAN — no packets will be sent")
		fmt.Printf("  dialers:      %d\n", len(dialers))
		fmt.Printf("  dials each:   %d\n", opts.Dials)
		fmt.Printf("  total dials:  %d\n", total)
		fmt.Printf("  jitter:       %d–%d ms\n",
			opts.JitterMin.Milliseconds(), opts.JitterMax.Milliseconds())
		fmt.Printf("  target:       %s\n", opts.Target)
		fmt.Printf("  est. duration: ~%ds\n",
			(total*(int(opts.JitterMin.Milliseconds())+int(opts.JitterMax.Milliseconds()))/2)/1000)
		fmt.Printf("  est. bytes:   ~%d KB\n", total*150/1024)
		if opts.Prefer != "" {
			fmt.Printf("  preference:   %s\n", opts.Prefer)
		}
		if opts.MaxDials > 0 && total > opts.MaxDials {
			fmt.Printf("  ⚠ would exceed safety cap (%d)\n", opts.MaxDials)
		}
		return
	}

	results := runner.Run(ctx, dialers, opts)
	sorted := runner.SortByMedian(results)

	truthAnswer, _ := runner.Dial(ctx,
		config.Dialer{Name: "truth", Protocol: config.ProtoDoH,
			Address: "https://cloudflare-dns.com/dns-query"},
		opts.Target, opts.Timeout)

	truthIPs := truthAnswer.IPs

	verdicts := map[string]compare.Verdict{}
	for _, r := range results {
		observed := mostCommonIPs(r.IPs)
		verdicts[r.Dialer.Name+"|"+r.Dialer.Protocol] = compare.Compare(observed, truthIPs)
	}

	output.Render(sorted, verdicts, opts)

	if len(truthIPs) == 0 {
		fmt.Println("note: ground-truth dial failed — answer comparison skipped")
	}

	if opts.Tune {
		tune.Print(sorted, verdicts, opts)
	}
}

func mostCommonIPs(samples [][]string) []net.IP {
	counts := map[string]int{}
	raw := map[string][]net.IP{}
	for _, ips := range samples {
		key := joinSorted(ips)
		counts[key]++
		if _, ok := raw[key]; !ok {
			parsed := make([]net.IP, 0, len(ips))
			for _, s := range ips {
				if ip := net.ParseIP(s); ip != nil {
					parsed = append(parsed, ip)
				}
			}
			raw[key] = parsed
		}
	}
	var best string
	bestN := -1
	for k, n := range counts {
		if n > bestN {
			bestN, best = n, k
		}
	}
	return raw[best]
}

func joinSorted(s []string) string {
	cp := append([]string(nil), s...)
	sort.Strings(cp)
	return strings.Join(cp, ",")
}

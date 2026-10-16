package output

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

func color(ms int64) string {
	switch {
	case ms < 30:
		return green
	case ms < 80:
		return yellow
	default:
		return red
	}
}

func verdictColor(v compare.Verdict) string {
	switch v.Label() {
	case "match":
		return green
	case "partial":
		return yellow
	case "mismatch":
		return red + bold
	default:
		return dim
	}
}

func Render(results []runner.Result, verdicts map[string]compare.Verdict, opts config.Options) {
	tag := ""
	if opts.Tag != "" {
		tag = " [" + opts.Tag + "]"
	}
	fmt.Printf("%s%s IDT — Internet Dialer Test%s%s  %s(%d dials · target %s · jitter %d-%dms)%s\n\n",
		bold, cyan, tag, reset,
		dim, opts.Dials, opts.Target,
		opts.JitterMin.Milliseconds(), opts.JitterMax.Milliseconds(), reset)

	fmt.Printf("%s%-12s %-9s %-9s %-9s %-9s %-8s %-7s %-9s %s%s\n",
		bold, "Provider", "Protocol", "Median", "P90", "Jitter",
		"OK", "Status", "Answers", "Detail", reset)
	fmt.Println(dim + strings.Repeat("─", 108) + reset)

	for _, r := range results {
		key := r.Dialer.Name + "|" + r.Dialer.Protocol
		v := verdicts[key]

		if r.Summary.Count == 0 {
			fmt.Printf("%-12s %-9s %s%-9s%s %-9s %-9s %-8s %s%-7s%s %-9s %s\n",
				r.Dialer.Name, r.Dialer.Protocol,
				red, "-", reset, "-", "-",
				fmt.Sprintf("0/%d", r.Total), red, "❌ dead", reset,
				v.Label(), "")
			continue
		}

		ms := r.Summary.Median.Milliseconds()
		c := color(ms)
		status := "✅"
		if r.Summary.Count < r.Total {
			status = "⚠️ "
		}
		vc := verdictColor(v)

		fmt.Printf("%-12s %-9s %s%-9s%s %-9s %-9s %-8s %s%-7s%s %s%-9s%s %s\n",
			r.Dialer.Name, r.Dialer.Protocol,
			c, fmt.Sprintf("%d ms", ms), reset,
			fmt.Sprintf("%d ms", r.Summary.P90.Milliseconds()),
			fmt.Sprintf("±%d", r.Summary.Jitter.Milliseconds()),
			fmt.Sprintf("%d/%d", r.Summary.Count, r.Total),
			c, status, reset,
			vc, v.Label(), reset,
			dim+v.Detail()+reset)
	}
	fmt.Println()
}

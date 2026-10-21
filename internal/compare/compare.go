package compare

import (
	"net"
	"sort"
	"strings"
)

// Truth is the reference IP set collected from a trusted encrypted dialer.
type Truth struct {
	IPs    []net.IP
	Source string
}

// Verdict is the comparison result for one dialer.
type Verdict struct {
	Checked    bool
	Matches    int
	Total      int
	SampleGot  []string
	SampleWant []string
}

// Compare checks whether observed IPs overlap the truth set.
// Partial matches count as "partial". No overlap = "mismatch".
func Compare(observed, truth []net.IP) Verdict {
	v := Verdict{Checked: true}
	if len(truth) == 0 || len(observed) == 0 {
		return v
	}

	// Build reference map for truth set with normalized IP strings
	want := make(map[string]bool)
	for _, ip := range truth {
		if normalized := normalizeIP(ip); normalized != "" {
			want[normalized] = true
		}
	}

	v.SampleWant = ipStrings(truth)

	// Deduplicate observed IPs to prevent skewed totals from repeated answers
	seenObserved := make(map[string]bool)
	var uniqueObserved []net.IP
	for _, ip := range observed {
		norm := normalizeIP(ip)
		if norm != "" && !seenObserved[norm] {
			seenObserved[norm] = true
			uniqueObserved = append(uniqueObserved, ip)
		}
	}

	for _, ip := range uniqueObserved {
		v.Total++
		norm := normalizeIP(ip)
		if want[norm] {
			v.Matches++
		}
	}

	v.SampleGot = ipStrings(uniqueObserved)
	return v
}

// normalizeIP standardizes IP string representation to handle IPv4/IPv6 nuances safely.
func normalizeIP(ip net.IP) string {
	if ip == nil {
		return ""
	}
	if ip4 := ip.To4(); ip4 != nil {
		return ip4.String()
	}
	return ip.String()
}

func ipStrings(ips []net.IP) []string {
	seen := make(map[string]bool)
	var out []string
	for _, ip := range ips {
		norm := normalizeIP(ip)
		if norm != "" && !seen[norm] {
			seen[norm] = true
			out = append(out, norm)
		}
	}
	sort.Strings(out)
	return out
}

func (v Verdict) Label() string {
	if !v.Checked {
		return "—"
	}
	if v.Total == 0 {
		return "?"
	}
	if v.Matches == v.Total {
		return "match"
	}
	if v.Matches > 0 {
		return "partial"
	}
	return "mismatch"
}

func (v Verdict) Detail() string {
	if !v.Checked || v.Total == 0 {
		return ""
	}
	return strings.Join(v.SampleGot, ",")
}

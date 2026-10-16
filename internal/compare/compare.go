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
	want := map[string]bool{}
	for _, ip := range truth {
		want[ip.String()] = true
	}
	for _, ip := range observed {
		v.Total++
		if want[ip.String()] {
			v.Matches++
		}
	}
	v.SampleGot = ipStrings(observed)
	v.SampleWant = ipStrings(truth)
	return v
}

func ipStrings(ips []net.IP) []string {
	out := make([]string, 0, len(ips))
	for _, ip := range ips {
		out = append(out, ip.String())
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

package config

import "time"

const (
	ProtoStandard = "Standard"
	ProtoDoT      = "DoT"
	ProtoDoH      = "DoH"
)

// Dialer is a public DNS endpoint IDT can dial.
type Dialer struct {
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Address  string `json:"address"`
}

// Options carries runtime configuration.
type Options struct {
	Target     string
	Dials      int
	Timeout    time.Duration
	JitterMin  time.Duration
	JitterMax  time.Duration
	MaxDials   int
	Plan       bool
	Safety     bool
	Check      bool
	Tune       bool
	Prefer     string
	Tag        string
}

// SafeTargets is a rotation pool. All neutral. None flaggable.
var SafeTargets = []string{
	"example.com",
	"example.org",
	"example.net",
	"iana.org",
	"www.iana.org",
}

func DefaultDialers() []Dialer {
	return []Dialer{
		{"Cloudflare", ProtoDoH, "https://cloudflare-dns.com/dns-query"},
		{"Cloudflare", ProtoDoT, "one.one.one.one:853"},
		{"Cloudflare", ProtoStandard, "1.1.1.1:53"},

		{"Google", ProtoDoH, "https://dns.google/dns-query"},
		{"Google", ProtoDoT, "dns.google:853"},
		{"Google", ProtoStandard, "8.8.8.8:53"},

		{"Quad9", ProtoDoH, "https://dns.quad9.net/dns-query"},
		{"Quad9", ProtoDoT, "dns.quad9.net:853"},
		{"Quad9", ProtoStandard, "9.9.9.9:53"},

		{"OpenDNS", ProtoDoH, "https://doh.opendns.com/dns-query"},
		{"OpenDNS", ProtoDoT, "dns.opendns.com:853"},
		{"OpenDNS", ProtoStandard, "208.67.222.222:53"},

		{"AdGuard", ProtoDoH, "https://dns.adguard-dns.com/dns-query"},
		{"AdGuard", ProtoDoT, "dns.adguard-dns.com:853"},
		{"AdGuard", ProtoStandard, "94.140.14.14:53"},

		{"Mullvad", ProtoDoH, "https://dns.mullvad.net/dns-query"},
		{"Mullvad", ProtoDoT, "dns.mullvad.net:853"},
		{"Mullvad", ProtoStandard, "194.242.2.2:53"},

		{"CIRA", ProtoDoH, "https://protected.canadianshield.cira.ca/dns-query"},
		{"CIRA", ProtoDoT, "protected.canadianshield.cira.ca:853"},
		{"CIRA", ProtoStandard, "149.112.121.10:53"},

		{"ControlD", ProtoDoH, "https://freedns.controld.com/p0"},
		{"ControlD", ProtoDoT, "p0.freedns.controld.com:853"},
		{"ControlD", ProtoStandard, "76.76.2.0:53"},
	}
}

func DefaultOptions() Options {
	return Options{
		Target:    "example.com",
		Dials:     3,
		Timeout:   3 * time.Second,
		JitterMin: 300 * time.Millisecond,
		JitterMax: 900 * time.Millisecond,
		MaxDials:  90,
	}
}

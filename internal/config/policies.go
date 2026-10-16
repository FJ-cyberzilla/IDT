package config

// Policy captures non-network facts the tool cannot measure:
// privacy stance, filtering behavior, jurisdiction, and log retention.
// Values are the operator's own published claims — verify before trusting.

type Policy struct {
	Name         string
	PrivacyScore int    // 1 (worst) to 5 (best)
	Blocking     string // "none" | "malware" | "ads" | "malware+ads"
	Jurisdiction string // ISO 3166-1 alpha-2
	LogDays      int    // -1 unknown, 0 none
	Notes        string
	AndroidDoT   string // hostname to paste into Android Private DNS
}

var Policies = map[string]Policy{
	"Cloudflare": {
		Name: "Cloudflare", PrivacyScore: 4, Blocking: "none",
		Jurisdiction: "US", LogDays: 1,
		Notes:      "24h anonymized logs, no profiling",
		AndroidDoT: "one.one.one.one",
	},
	"Google": {
		Name: "Google", PrivacyScore: 3, Blocking: "none",
		Jurisdiction: "US", LogDays: 2,
		Notes:      "24-48h anonymized, US jurisdiction",
		AndroidDoT: "dns.google",
	},
	"Quad9": {
		Name: "Quad9", PrivacyScore: 5, Blocking: "malware",
		Jurisdiction: "CH", LogDays: 0,
		Notes:      "Swiss foundation, blocks malware, no logs",
		AndroidDoT: "dns.quad9.net",
	},
	"OpenDNS": {
		Name: "OpenDNS", PrivacyScore: 3, Blocking: "malware",
		Jurisdiction: "US", LogDays: 30,
		Notes:      "Cisco-owned, phishing + content filtering",
		AndroidDoT: "dns.opendns.com",
	},
	"AdGuard": {
		Name: "AdGuard", PrivacyScore: 4, Blocking: "malware+ads",
		Jurisdiction: "CY", LogDays: 0,
		Notes:      "Ad-blocking resolver, Cyprus-based",
		AndroidDoT: "dns.adguard-dns.com",
	},
	"Mullvad": {
		Name: "Mullvad", PrivacyScore: 5, Blocking: "none",
		Jurisdiction: "SE", LogDays: 0,
		Notes:      "Strict no-logs, Sweden, operator of Mullvad VPN",
		AndroidDoT: "dns.mullvad.net",
	},
	"CIRA": {
		Name: "CIRA", PrivacyScore: 5, Blocking: "malware",
		Jurisdiction: "CA", LogDays: 0,
		Notes:      "Canadian Shield, malware blocking, no logs",
		AndroidDoT: "protected.canadianshield.cira.ca",
	},
	"ControlD": {
		Name: "ControlD", PrivacyScore: 4, Blocking: "none",
		Jurisdiction: "CA", LogDays: 0,
		Notes:      "Policy tiers (p0 = unfiltered), Canada-based",
		AndroidDoT: "p0.freedns.controld.com",
	},
}

// FilterByPreference narrows the dialer list before running.
func FilterByPreference(dialers []Dialer, prefer string) []Dialer {
	if prefer == "" {
		return dialers
	}

	allowed := map[string]bool{}
	for name, p := range Policies {
		ok := false
		switch prefer {
		case "privacy":
			ok = p.PrivacyScore >= 4
		case "no-logging":
			ok = p.LogDays == 0
		case "blocking":
			ok = p.Blocking != "none"
		case "unfiltered":
			ok = p.Blocking == "none"
		case "swiss":
			ok = p.Jurisdiction == "CH"
		case "canadian":
			ok = p.Jurisdiction == "CA"
		}
		if ok {
			allowed[name] = true
		}
	}
	if len(allowed) == 0 {
		return dialers
	}

	var out []Dialer
	for _, d := range dialers {
		if allowed[d.Name] {
			out = append(out, d)
		}
	}
	return out
}

package configs

import (
	"encoding/json"
	"time"
)

// AppConfig defines the application-wide best performance and reliability settings.
type AppConfig struct {
	Version          string        `json:"version"`
	DefaultTarget    string        `json:"default_target"`
	FallbackTargets  []string      `json:"fallback_targets"`
	DialsPerEndpoint int           `json:"dials_per_endpoint"`
	TimeoutMS        int           `json:"timeout_ms"`
	JitterMinMS      int           `json:"jitter_min_ms"`
	JitterMaxMS      int           `json:"jitter_max_ms"`
	MaxConcurrency   int           `json:"max_concurrency"`
	RetryAttempts    int           `json:"retry_attempts"`
	PoolMaxIdle      int           `json:"pool_max_idle"`
	PoolIdleTimeout  time.Duration `json:"pool_idle_timeout"`
}

// NetworkDialer represents a curated, high-performance public DNS endpoint.
type NetworkDialer struct {
	Name         string `json:"name"`
	Protocol     string `json:"protocol"` // DoH, DoT, Standard
	Address      string `json:"address"`
	PrivacyScore int    `json:"privacy_score"` // 1-5
	Jurisdiction string `json:"jurisdiction"`
	ZeroLogs     bool   `json:"zero_logs"`
	MalwareBlock bool   `json:"malware_block"`
}

// MasterConfig combines AppConfig and curated NetworkDialers for IDT.
type MasterConfig struct {
	App     AppConfig       `json:"options"`
	Dialers []NetworkDialer `json:"dialers"`
}

// DefaultMasterConfig returns the ultimate production-grade default configuration
// optimized for speed, reliability, privacy, and low latency across global networks.
func DefaultMasterConfig() MasterConfig {
	return MasterConfig{
		App: AppConfig{
			Version:          "1.0.0",
			DefaultTarget:    "example.com",
			FallbackTargets:  []string{"example.com", "example.org", "example.net", "iana.org"},
			DialsPerEndpoint: 3,
			TimeoutMS:        2500,
			JitterMinMS:      200,
			JitterMaxMS:      600,
			MaxConcurrency:   32,
			RetryAttempts:    2,
			PoolMaxIdle:      64,
			PoolIdleTimeout:  90 * time.Second,
		},
		Dialers: []NetworkDialer{
			{
				Name: "Cloudflare", Protocol: "DoH", Address: "https://cloudflare-dns.com/dns-query",
				PrivacyScore: 4, Jurisdiction: "US", ZeroLogs: true, MalwareBlock: false,
			},
			{
				Name: "Cloudflare", Protocol: "DoT", Address: "one.one.one.one:853",
				PrivacyScore: 4, Jurisdiction: "US", ZeroLogs: true, MalwareBlock: false,
			},
			{
				Name: "Quad9", Protocol: "DoH", Address: "https://dns.quad9.net/dns-query",
				PrivacyScore: 5, Jurisdiction: "CH", ZeroLogs: true, MalwareBlock: true,
			},
			{
				Name: "Quad9", Protocol: "DoT", Address: "dns.quad9.net:853",
				PrivacyScore: 5, Jurisdiction: "CH", ZeroLogs: true, MalwareBlock: true,
			},
			{
				Name: "Mullvad", Protocol: "DoH", Address: "https://dns.mullvad.net/dns-query",
				PrivacyScore: 5, Jurisdiction: "SE", ZeroLogs: true, MalwareBlock: false,
			},
			{
				Name: "Mullvad", Protocol: "DoT", Address: "dns.mullvad.net:853",
				PrivacyScore: 5, Jurisdiction: "SE", ZeroLogs: true, MalwareBlock: false,
			},
			{
				Name: "AdGuard", Protocol: "DoH", Address: "https://dns.adguard-dns.com/dns-query",
				PrivacyScore: 4, Jurisdiction: "CY", ZeroLogs: true, MalwareBlock: true,
			},
			{
				Name: "CIRA", Protocol: "DoH", Address: "https://protected.canadianshield.cira.ca/dns-query",
				PrivacyScore: 5, Jurisdiction: "CA", ZeroLogs: true, MalwareBlock: true,
			},
		},
	}
}

// ToJSON serializes the master configuration into formatted JSON bytes.
func (mc MasterConfig) ToJSON() ([]byte, error) {
	return json.MarshalIndent(mc, "", "  ")
}

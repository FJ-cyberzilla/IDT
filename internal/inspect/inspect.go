package inspect

import (
	"context"
	"idt/internal/config"
	"idt/internal/guardcat"
)

// SystemDiagnostics aggregates all system, Termux, Android VPN, Private DNS, Browser DoH, Country, and Guardcat awareness data.
type SystemDiagnostics struct {
	DNSConfig           config.SystemDNSConfig `json:"dns_config"`
	InternetOnline      bool                   `json:"internet_online"`
	ConnectivityStatus  string                 `json:"connectivity_status"`
	GuardcatRecommendation string              `json:"guardcat_recommendation"`
}

// RunDiagnostics performs full system telemetry and diagnostics inspection.
func RunDiagnostics(ctx context.Context) SystemDiagnostics {
	dnsCfg := config.DiscoverSystemDNS()

	gc := guardcat.New(guardcat.DefaultConfig)
	online, connStatus := gc.CheckInternetConnectivity(ctx)
	rec := gc.RecommendAction()

	return SystemDiagnostics{
		DNSConfig:              dnsCfg,
		InternetOnline:         online,
		ConnectivityStatus:     connStatus,
		GuardcatRecommendation: rec,
	}
}

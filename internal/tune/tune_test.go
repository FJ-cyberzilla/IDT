package tune

import (
	"testing"
	"time"

	"idt/internal/compare"
	"idt/internal/config"
	"idt/internal/runner"
	"idt/internal/stats"
)

func TestFilterEligible(t *testing.T) {
	results := []runner.Result{
		{
			Dialer:  config.Dialer{Name: "Cloudflare", Protocol: config.ProtoDoH, Address: "https://..."},
			Summary: stats.Summary{Count: 3, Median: 20 * time.Millisecond},
			Total:   3,
		},
		{
			Dialer:  config.Dialer{Name: "Google", Protocol: config.ProtoStandard, Address: "8.8.8.8:53"},
			Summary: stats.Summary{Count: 3, Median: 10 * time.Millisecond},
			Total:   3,
		},
	}

	verdicts := map[string]compare.Verdict{}
	eligible := filterEligible(results, verdicts)

	// Standard protocol should be filtered out. Only DoH should remain.
	if len(eligible) != 1 {
		t.Fatalf("expected 1 eligible dialer, got %d", len(eligible))
	}
	if eligible[0].Result.Dialer.Name != "Cloudflare" {
		t.Errorf("expected Cloudflare, got %s", eligible[0].Result.Dialer.Name)
	}
}

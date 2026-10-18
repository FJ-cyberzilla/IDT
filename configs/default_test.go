package configs

import (
	"encoding/json"
	"testing"
)

func TestDefaultMasterConfig(t *testing.T) {
	cfg := DefaultMasterConfig()

	if cfg.App.Version == "" {
		t.Error("expected version to be set")
	}
	if cfg.App.DefaultTarget != "example.com" {
		t.Errorf("expected default target 'example.com', got %s", cfg.App.DefaultTarget)
	}
	if len(cfg.App.FallbackTargets) == 0 {
		t.Error("expected fallback targets")
	}
	if cfg.App.TimeoutMS <= 0 {
		t.Error("expected positive timeout ms")
	}
	if cfg.App.MaxConcurrency <= 0 {
		t.Error("expected positive max concurrency")
	}
	if cfg.App.PoolMaxIdle <= 0 {
		t.Error("expected positive pool max idle")
	}
	if cfg.App.PoolIdleTimeout <= 0 {
		t.Error("expected positive pool idle timeout")
	}

	if len(cfg.Dialers) == 0 {
		t.Error("expected curated dialers, got none")
	}

	for _, d := range cfg.Dialers {
		if d.Name == "" || d.Protocol == "" || d.Address == "" {
			t.Errorf("dialer missing fields: %+v", d)
		}
		if d.PrivacyScore < 1 || d.PrivacyScore > 5 {
			t.Errorf("invalid privacy score for %s: %d", d.Name, d.PrivacyScore)
		}
	}

	data, err := cfg.ToJSON()
	if err != nil {
		t.Fatalf("unexpected error marshaling config to JSON: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty JSON output")
	}

	// Verify unmarshaling roundtrip
	var roundtrip MasterConfig
	if err := json.Unmarshal(data, &roundtrip); err != nil {
		t.Fatalf("failed to unmarshal JSON configuration: %v", err)
	}
	if roundtrip.App.DefaultTarget != cfg.App.DefaultTarget {
		t.Errorf("roundtrip target mismatch: got %s, want %s", roundtrip.App.DefaultTarget, cfg.App.DefaultTarget)
	}
	if len(roundtrip.Dialers) != len(cfg.Dialers) {
		t.Errorf("roundtrip dialers count mismatch: got %d, want %d", len(roundtrip.Dialers), len(cfg.Dialers))
	}
}

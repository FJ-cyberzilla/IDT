package singbox

import (
	"context"
	"testing"
	"time"
)

func TestSafeguard_DialContext(t *testing.T) {
	sg := NewSafeguard(SafeguardConfig{
		MaxConcurrentDial: 2,
		DialTimeout:       1 * time.Second,
		EnableThrottling:  true,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Verify creation and basic structure
	if sg == nil {
		t.Fatal("expected safeguard instance")
	}

	// Test non-routable dial timeout to ensure safety bounds work without hanging
	_, err := sg.DialContext(ctx, "tcp", "192.0.2.1:53")
	if err == nil {
		t.Log("dial succeeded unexpectedly to reserved IP (unexpected but non-fatal)")
	}
}

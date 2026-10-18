package vault

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLivePoolWarmupAndPick(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "datalist_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write a candidate with an invalid/unreachable IP so it gets dropped cleanly or tested
	csvContent := "ip_address,name\n192.0.2.1,dead-resolver.example.com\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "test.csv"), []byte(csvContent), 0644); err != nil {
		t.Fatalf("failed to write csv: %v", err)
	}

	logger := NoopLogger{}
	pool := NewLivePool(logger, tmpDir)

	ctx := context.Background()
	_ = pool.Warmup(ctx)

	// Since 192.0.2.1 is TEST-NET (unreachable), live pool should have 0 live dialers, falling back to fallbacks
	d := pool.Pick()
	if d.Name == "" {
		t.Errorf("expected fallback dialer when live pool is empty, got empty")
	}
}

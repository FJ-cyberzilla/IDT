package vault

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestJSONLFileLogger(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vault_test_runs_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logger, err := NewJSONLFileLogger(tmpDir)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	logger.LogWarmup("1.1.1.1:53", "Cloudflare", "ok", 10*time.Millisecond, "", "")
	logger.LogTest("Cloudflare", "example.com", "tencent", "ok", []string{"93.184.216.34"}, []string{"93.184.216.34"}, 15*time.Millisecond, "")
	logger.LogSnapshot(54, 30, 25, map[string]int{"timeout": 10, "mismatch": 5})

	if err := logger.Close(); err != nil {
		t.Fatalf("failed to close logger: %v", err)
	}

	matches, err := filepath.Glob(filepath.Join(tmpDir, "run_*.jsonl"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("expected 1 jsonl file, got %d (err: %v)", len(matches), err)
	}
}

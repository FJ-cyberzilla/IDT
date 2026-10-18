package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCandidates(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "datalist_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	csvContent := "ip_address,name,country_code\n1.2.3.4,ns1.example.at.,AT\n5.6.7.8,,US\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "test.csv"), []byte(csvContent), 0644); err != nil {
		t.Fatalf("failed to write csv: %v", err)
	}

	dialers, err := LoadCandidates(tmpDir)
	if err != nil {
		t.Fatalf("LoadCandidates failed: %v", err)
	}

	if len(dialers) != 2 {
		t.Fatalf("expected 2 dialers, got %d", len(dialers))
	}

	if dialers[0].Address != "1.2.3.4:53" || dialers[0].Name != "ns1.example.at" {
		t.Errorf("unexpected first dialer: %+v", dialers[0])
	}
	if dialers[1].Address != "5.6.7.8:53" || dialers[1].Name != "5.6.7.8" {
		t.Errorf("unexpected second dialer: %+v", dialers[1])
	}
}

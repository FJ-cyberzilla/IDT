package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyConfigHMAC_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "servers.json")
	configData := []byte(`{"options": {"target": "example.com"}, "dialers": []}`)

	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	// First call should auto-generate signature file (.sig)
	if err := VerifyConfigHMAC(configPath, configData, "test-key"); err != nil {
		t.Errorf("expected success on first verify/generate, got: %v", err)
	}

	// Second call should verify successfully against existing .sig file
	if err := VerifyConfigHMAC(configPath, configData, "test-key"); err != nil {
		t.Errorf("expected success on second verify, got: %v", err)
	}
}

func TestVerifyConfigHMAC_Tampered(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "servers.json")
	originalData := []byte(`{"options": {"target": "example.com"}, "dialers": []}`)
	tamperedData := []byte(`{"options": {"target": "evil.com"}, "dialers": []}`)

	if err := os.WriteFile(configPath, originalData, 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	// Generate valid signature for original data
	if err := VerifyConfigHMAC(configPath, originalData, "test-key"); err != nil {
		t.Fatalf("failed to generate initial signature: %v", err)
	}

	// Overwrite config with tampered data
	if err := os.WriteFile(configPath, tamperedData, 0644); err != nil {
		t.Fatalf("failed to write tampered config: %v", err)
	}

	// Verification should fail because data does not match stored signature
	if err := VerifyConfigHMAC(configPath, tamperedData, "test-key"); err == nil {
		t.Error("expected HMAC verification to fail for tampered config, but it succeeded")
	}
}

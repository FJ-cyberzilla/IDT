package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEncryptDatalistCSVs(t *testing.T) {
	datalistDir := "../../datalist"
	matches, err := filepath.Glob(filepath.Join(datalistDir, "*.csv"))
	if err != nil || len(matches) == 0 {
		t.Skip("datalist dir not found or glob failed")
		return
	}

	for _, csvPath := range matches {
		data, err := os.ReadFile(csvPath)
		if err != nil {
			continue
		}

		encrypted, err := Encrypt(data, "")
		if err != nil {
			t.Fatalf("failed to encrypt %s: %v", csvPath, err)
		}

		encPath := strings.TrimSuffix(csvPath, ".csv") + ".enc"
		if err := os.WriteFile(encPath, encrypted, 0644); err != nil {
			t.Fatalf("failed to write %s: %v", encPath, err)
		}

		// Remove plaintext csv
		_ = os.Remove(csvPath)
	}
}

package vault

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"idt/internal/config"
)

// Candidate represents a resolver candidate loaded from seed CSVs.
type Candidate struct {
	IP          string
	Name        string
	ASOrg       string
	Country     string
	Reliability float64
	Dialer      config.Dialer
}

// LoadCandidates loads all CSV or encrypted (.enc) files in the given datalist directory and returns candidate Dialers.
func LoadCandidates(datalistDir string) ([]config.Dialer, error) {
	if datalistDir == "" {
		datalistDir = "datalist"
	}

	// Match both .csv and .enc files
	patterns := []string{
		filepath.Join(datalistDir, "*.csv"),
		filepath.Join(datalistDir, "*.enc"),
	}

	var matches []string
	for _, p := range patterns {
		if found, err := filepath.Glob(p); err == nil {
			matches = append(matches, found...)
		}
	}

	var dialers []config.Dialer
	seenIPs := make(map[string]bool)

	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var csvBytes []byte
		if strings.HasSuffix(strings.ToLower(path), ".enc") {
			decrypted, err := Decrypt(data, "")
			if err != nil {
				// Try falling back or skip if decryption fails
				continue
			}
			csvBytes = decrypted
		} else {
			csvBytes = data
		}

		reader := csv.NewReader(bytes.NewReader(csvBytes))
		// Read header
		header, err := reader.Read()
		if err != nil {
			continue
		}

		// Find column indices
		ipIdx := -1
		nameIdx := -1
		orgIdx := -1

		for i, col := range header {
			colTrim := strings.TrimSpace(strings.ToLower(col))
			switch colTrim {
			case "ip_address", "ip":
				ipIdx = i
			case "name", "hostname":
				nameIdx = i
			case "as_org", "org":
				orgIdx = i
			}
		}

		if ipIdx == -1 {
			continue
		}

		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}

			if ipIdx >= len(record) {
				continue
			}

			ip := strings.TrimSpace(record[ipIdx])
			if ip == "" || seenIPs[ip] {
				continue
			}

			name := ""
			if nameIdx != -1 && nameIdx < len(record) {
				name = strings.TrimSpace(record[nameIdx])
			}
			if name == "" && orgIdx != -1 && orgIdx < len(record) {
				name = strings.TrimSpace(record[orgIdx])
			}
			if name == "" {
				name = ip
			}

			// Clean trailing dot if present
			name = strings.TrimSuffix(name, ".")

			seenIPs[ip] = true
			addr := fmt.Sprintf("%s:53", ip)
			dialers = append(dialers, config.Dialer{
				Name:     name,
				Protocol: config.ProtoStandard,
				Address:  addr,
			})
		}
	}

	return dialers, nil
}

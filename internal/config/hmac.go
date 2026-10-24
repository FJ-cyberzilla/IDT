package config

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"strings"
)

const defaultHMACKey = "idt-secure-config-hmac-master-key-2026"

// ComputeHMAC computes HMAC-SHA256 of data using the given key.
func ComputeHMAC(data []byte, key string) string {
	if key == "" {
		key = defaultHMACKey
	}
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyConfigHMAC verifies the HMAC-SHA256 signature of the config file at configPath.
// It checks for a companion signature file (configPath + ".sig").
// If the signature file does not exist, it generates and writes one (packaging / first boot helper).
// If the signature file exists, it validates the signature against the current file data.
// Returns an error if verification fails (e.g. tampering detected).
func VerifyConfigHMAC(configPath string, data []byte, key string) error {
	if key == "" {
		key = defaultHMACKey
	}
	sigPath := configPath + ".sig"

	expectedSig := ComputeHMAC(data, key)

	sigBytes, err := os.ReadFile(sigPath)
	if err != nil {
		// If signature file does not exist, initialize it (packaging helper).
		if os.IsNotExist(err) {
			if writeErr := os.WriteFile(sigPath, []byte(expectedSig), 0644); writeErr != nil {
				return writeErr
			}
			return nil
		}
		return err
	}

	actualSig := strings.TrimSpace(string(sigBytes))
	if !hmac.Equal([]byte(actualSig), []byte(expectedSig)) {
		return errors.New("HMAC signature verification failed: configuration file has been tampered with")
	}

	return nil
}

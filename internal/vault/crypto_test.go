package vault

import (
	"bytes"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	plaintext := []byte("ip_address,name\n1.2.3.4,test.at\n")
	ciphertext, err := Encrypt(plaintext, "test-secret")
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	decrypted, err := Decrypt(ciphertext, "test-secret")
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("decrypted text does not match plaintext: got %s, want %s", decrypted, plaintext)
	}

	secondCiphertext, err := Encrypt(plaintext, "test-secret")
	if err != nil {
		t.Fatalf("second Encrypt failed: %v", err)
	}
	saltStart := len(ciphertextVersion)
	saltEnd := saltStart + saltSize
	if bytes.Equal(ciphertext[saltStart:saltEnd], secondCiphertext[saltStart:saltEnd]) {
		t.Error("Encrypt reused the same salt")
	}

	if _, err := Decrypt(ciphertext, "wrong-secret"); err == nil {
		t.Error("Decrypt succeeded with the wrong passphrase")
	}
}

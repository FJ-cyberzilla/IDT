package vault

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
)

// defaultPassphrase is the default master secret used to encrypt/decrypt datalist seed files.
const defaultPassphrase = "idt-secure-vault-resolver-seed-key-2026"

const (
	ciphertextVersion = "IDT\x01"
	saltSize          = 16
	keySize           = 32
	pbkdf2Iterations  = 600000
)

// deriveKey generates a 32-byte AES key from a passphrase using PBKDF2.
func deriveKey(passphrase string, salt []byte) ([]byte, error) {
	if passphrase == "" {
		passphrase = defaultPassphrase
	}
	return pbkdf2.Key(sha256.New, passphrase, salt, pbkdf2Iterations, keySize)
}

// Encrypt encrypts plaintext using AES-256-GCM with the given passphrase.
func Encrypt(plaintext []byte, passphrase string) ([]byte, error) {
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	key, err := deriveKey(passphrase, salt)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := make([]byte, 0, len(ciphertextVersion)+len(salt)+len(nonce)+len(plaintext)+gcm.Overhead())
	ciphertext = append(ciphertext, ciphertextVersion...)
	ciphertext = append(ciphertext, salt...)
	ciphertext = append(ciphertext, nonce...)
	ciphertext = gcm.Seal(ciphertext, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts ciphertext using AES-256-GCM with the given passphrase.
func Decrypt(ciphertext []byte, passphrase string) ([]byte, error) {
	if len(ciphertext) < len(ciphertextVersion)+saltSize || !bytes.Equal(ciphertext[:len(ciphertextVersion)], []byte(ciphertextVersion)) {
		return nil, errors.New("invalid ciphertext format")
	}
	offset := len(ciphertextVersion)
	salt := ciphertext[offset : offset+saltSize]

	key, err := deriveKey(passphrase, salt)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	offset += saltSize
	if len(ciphertext)-offset < gcm.NonceSize()+gcm.Overhead() {
		return nil, errors.New("ciphertext too short")
	}

	nonce := ciphertext[offset : offset+gcm.NonceSize()]
	cipherBytes := ciphertext[offset+gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, cipherBytes, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
)

// GetEncryptionKey returns the encryption key from Settings or generates a new one
// Note: This function assumes the caller already holds the mutex lock if needed
func GetEncryptionKey() []byte {
	settings := GetSettings()

	// If key exists in settings, decode and return it
	if settings.EncryptionKey != "" {
		if key, err := hex.DecodeString(settings.EncryptionKey); err == nil && len(key) == 32 {
			return key
		}
	}

	// Generate new key
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		// Fallback to a deterministic key if random generation fails
		fallback := "godbadmin-fallback-key-32bytes!"
		copy(key, []byte(fallback))
		return key[:32]
	}

	// Store key in settings
	settings.EncryptionKey = hex.EncodeToString(key)

	return key
}

// Encrypt encrypts plain text using AES-256-GCM
func Encrypt(plaintext string) (string, error) {
	key := GetEncryptionKey()

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts cipher text using AES-256-GCM
func Decrypt(ciphertext string) (string, error) {
	key := GetEncryptionKey()

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, cipherData := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

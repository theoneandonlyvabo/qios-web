package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

var (
	ErrKeyNotSet      = errors.New("encryption: key not set")
	ErrInvalidKey     = errors.New("encryption: key must be 32 bytes")
	ErrInvalidPayload = errors.New("encryption: invalid ciphertext")
)

var encryptionKey []byte

// Init loads the AES-256 key from a 64-char hex string. Call once at startup.
func Init(keyHex string) error {
	if keyHex == "" {
		return ErrKeyNotSet
	}
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return fmt.Errorf("encryption: invalid hex key: %w", err)
	}
	if len(key) != 32 {
		return ErrInvalidKey
	}
	encryptionKey = key
	return nil
}

// Encrypt returns hex(nonce + ciphertext) using AES-256-GCM.
func Encrypt(plaintext string) (string, error) {
	if len(encryptionKey) == 0 {
		return "", ErrKeyNotSet
	}
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("encryption: new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("encryption: new gcm: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("encryption: generate nonce: %w", err)
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(sealed), nil
}

// Decrypt reverses Encrypt.
func Decrypt(ciphertextHex string) (string, error) {
	if len(encryptionKey) == 0 {
		return "", ErrKeyNotSet
	}
	data, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", fmt.Errorf("encryption: decode hex: %w", err)
	}
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("encryption: new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("encryption: new gcm: %w", err)
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrInvalidPayload
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("encryption: decrypt: %w", err)
	}
	return string(plaintext), nil
}

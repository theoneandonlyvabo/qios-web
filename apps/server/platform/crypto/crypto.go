// platform/crypto/crypto.go
//
// Enkripsi dan dekripsi data sensitif menggunakan AES-256-GCM.
// Dipakai untuk menyimpan xendit_secret_key di tabel businesses.
//
// Format ciphertext: nonce (12 bytes) + ciphertext (variable) — disimpan sebagai hex string.
// Key harus 32 bytes (256-bit) — dibaca dari ENCRYPTION_KEY env var via config.

package crypto

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
	ErrKeyNotSet      = errors.New("crypto: encryption key not set")
	ErrInvalidKey     = errors.New("crypto: encryption key must be 32 bytes")
	ErrInvalidPayload = errors.New("crypto: invalid ciphertext")
)

// encryptionKey adalah key aktif yang di-set saat startup via Init.
var encryptionKey []byte

// Init menyimpan encryption key ke package state.
// Dipanggil sekali di main setelah config di-load.
// keyHex harus 64 hex chars (= 32 bytes decoded).
func Init(keyHex string) error {
	if keyHex == "" {
		return ErrKeyNotSet
	}
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return fmt.Errorf("crypto: invalid hex key: %w", err)
	}
	if len(key) != 32 {
		return ErrInvalidKey
	}
	encryptionKey = key
	return nil
}

// Encrypt mengenkripsi plaintext string dan mengembalikan hex-encoded ciphertext.
// Format output: hex(nonce + ciphertext).
func Encrypt(plaintext string) (string, error) {
	if len(encryptionKey) == 0 {
		return "", ErrKeyNotSet
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("crypto: new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("crypto: new gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("crypto: generate nonce: %w", err)
	}

	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(sealed), nil
}

// Decrypt mendekripsi hex-encoded ciphertext yang dihasilkan Encrypt.
func Decrypt(ciphertextHex string) (string, error) {
	if len(encryptionKey) == 0 {
		return "", ErrKeyNotSet
	}

	data, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", fmt.Errorf("crypto: decode hex: %w", err)
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("crypto: new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("crypto: new gcm: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrInvalidPayload
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("crypto: decrypt: %w", err)
	}

	return string(plaintext), nil
}
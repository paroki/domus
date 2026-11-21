// Package crypt provides AES-256-GCM authenticated encryption at rest and
// HMAC-SHA256 blind indexing derived from a single APP_SECRET via HKDF-SHA256.
package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	// MinSecretLen is the minimum required length in bytes for APP_SECRET.
	MinSecretLen = 32
	// Version1 is the binary wire format version byte.
	Version1 byte = 0x01
	// StringPrefixV1 prefixes base64url-encoded v1 ciphertexts stored in text columns.
	StringPrefixV1 = "enc:v1:"

	hkdfSalt       = "domus:crypt:hkdf-salt:v1"
	hkdfInfoAESGCM = "domus:crypt:aes-gcm:v1"
	hkdfInfoBlind  = "domus:crypt:blind-index:v1"
	subkeyLen      = 32
)

var (
	// ErrSecretTooShort is returned when APP_SECRET is empty or shorter than MinSecretLen bytes.
	ErrSecretTooShort = fmt.Errorf("crypt: secret must be at least %d bytes", MinSecretLen)
	// ErrInvalidCiphertext is returned when a ciphertext payload is malformed or truncated.
	ErrInvalidCiphertext = errors.New("crypt: invalid or truncated ciphertext")
	// ErrUnsupportedVersion is returned when a ciphertext carries an unknown version byte.
	ErrUnsupportedVersion = errors.New("crypt: unsupported ciphertext version")
	// ErrDecryptFailed is returned when AES-GCM authentication fails.
	ErrDecryptFailed = errors.New("crypt: ciphertext authentication failed")

	version1AAD = []byte{Version1}
)

// Cipher provides AES-256-GCM authenticated encryption and HMAC-SHA256 blind indexing.
type Cipher struct {
	aead     cipher.AEAD
	blindKey []byte
}

// New derives isolated AES-256-GCM and HMAC-SHA256 subkeys from secret via HKDF-SHA256.
func New(secret string) (*Cipher, error) {
	trimmed := strings.TrimSpace(secret)
	if len(trimmed) < MinSecretLen {
		return nil, ErrSecretTooShort
	}

	salt := []byte(hkdfSalt)
	encKey, err := hkdf.Key(sha256.New, []byte(trimmed), salt, hkdfInfoAESGCM, subkeyLen)
	if err != nil {
		return nil, fmt.Errorf("crypt: derive encryption key: %w", err)
	}
	blindKey, err := hkdf.Key(sha256.New, []byte(trimmed), salt, hkdfInfoBlind, subkeyLen)
	if err != nil {
		return nil, fmt.Errorf("crypt: derive blind index key: %w", err)
	}

	block, err := aes.NewCipher(encKey)
	if err != nil {
		return nil, fmt.Errorf("crypt: init aes cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypt: init gcm: %w", err)
	}

	return &Cipher{
		aead:     aead,
		blindKey: blindKey,
	}, nil
}

// EncryptBytes encrypts plaintext (e.g. KTP photo bytes) using AES-256-GCM.
// Wire format: [Version1 (1B, authenticated via AAD) | Nonce (12B) | Ciphertext + Tag (16B)].
func (c *Cipher) EncryptBytes(plaintext []byte) ([]byte, error) {
	nonceSize := c.aead.NonceSize()
	out := make([]byte, 1+nonceSize, 1+nonceSize+len(plaintext)+c.aead.Overhead())
	out[0] = Version1
	nonce := out[1 : 1+nonceSize]
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("crypt: generate nonce: %w", err)
	}
	return c.aead.Seal(out, nonce, plaintext, version1AAD), nil
}

// DecryptBytes verifies and decrypts a versioned binary payload produced by EncryptBytes.
func (c *Cipher) DecryptBytes(ciphertext []byte) ([]byte, error) {
	nonceSize := c.aead.NonceSize()
	minLen := 1 + nonceSize + c.aead.Overhead()
	if len(ciphertext) < minLen {
		return nil, ErrInvalidCiphertext
	}
	if ciphertext[0] != Version1 {
		return nil, ErrUnsupportedVersion
	}
	nonce := ciphertext[1 : 1+nonceSize]
	sealed := ciphertext[1+nonceSize:]
	plain, err := c.aead.Open(nil, nonce, sealed, version1AAD)
	if err != nil {
		return nil, ErrDecryptFailed
	}
	return plain, nil
}

// EncryptString encrypts a UTF-8 string (e.g. LLM API key, NIK) into "enc:v1:<base64url>".
func (c *Cipher) EncryptString(plaintext string) (string, error) {
	raw, err := c.EncryptBytes([]byte(plaintext))
	if err != nil {
		return "", err
	}
	return StringPrefixV1 + base64.RawURLEncoding.EncodeToString(raw), nil
}

// DecryptString decrypts an "enc:v1:<base64url>" string produced by EncryptString.
func (c *Cipher) DecryptString(encoded string) (string, error) {
	if !IsEncryptedString(encoded) {
		return "", ErrInvalidCiphertext
	}
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(encoded, StringPrefixV1))
	if err != nil {
		return "", ErrInvalidCiphertext
	}
	plain, err := c.DecryptBytes(raw)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// IsEncryptedString reports whether s carries the v1 encrypted string prefix.
func IsEncryptedString(s string) bool {
	return strings.HasPrefix(s, StringPrefixV1)
}

// BlindIndex computes a deterministic 64-char hex HMAC-SHA256 digest over strings.TrimSpace(value)
// for exact-match DB lookups (e.g. NIK uniqueness). Returns "" if the trimmed value is empty.
func (c *Cipher) BlindIndex(value string) string {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return ""
	}
	mac := hmac.New(sha256.New, c.blindKey)
	mac.Write([]byte(normalized))
	return hex.EncodeToString(mac.Sum(nil))
}

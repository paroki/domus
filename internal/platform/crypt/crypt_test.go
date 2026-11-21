package crypt_test

import (
	"bytes"
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	"github.com/paroki/domus/internal/platform/crypt"
)

const testSecret = "0123456789abcdef0123456789abcdef"

func TestNewValidation(t *testing.T) {
	cases := []struct {
		name    string
		secret  string
		wantErr error
	}{
		{name: "empty", secret: "", wantErr: crypt.ErrSecretTooShort},
		{name: "whitespace only", secret: strings.Repeat(" ", 40), wantErr: crypt.ErrSecretTooShort},
		{name: "31 bytes", secret: strings.Repeat("a", 31), wantErr: crypt.ErrSecretTooShort},
		{name: "32 bytes with spaces trimmed to 31", secret: " " + strings.Repeat("a", 31) + " ", wantErr: crypt.ErrSecretTooShort},
		{name: "exact 32 bytes", secret: testSecret, wantErr: nil},
		{name: "64 hex chars", secret: testSecret + testSecret, wantErr: nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, err := crypt.New(tc.secret)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("New() err = %v, want %v", err, tc.wantErr)
				}
				if c != nil {
					t.Fatalf("New() returned non-nil Cipher on error")
				}
				return
			}
			if err != nil {
				t.Fatalf("New() unexpected err = %v", err)
			}
			if c == nil {
				t.Fatalf("New() returned nil Cipher")
			}
		})
	}
}

func TestEncryptDecryptString(t *testing.T) {
	c, err := crypt.New(testSecret)
	if err != nil {
		t.Fatalf("New() err = %v", err)
	}

	inputs := []string{
		"",
		"3171010101900001",
		"sk-ant-api03-very-secret-key-with-unicode-✓",
	}

	for _, plain := range inputs {
		enc1, err := c.EncryptString(plain)
		if err != nil {
			t.Fatalf("EncryptString(%q) err = %v", plain, err)
		}
		enc2, err := c.EncryptString(plain)
		if err != nil {
			t.Fatalf("EncryptString(%q) second call err = %v", plain, err)
		}
		if !crypt.IsEncryptedString(enc1) {
			t.Fatalf("IsEncryptedString(%q) = false, want true", enc1)
		}
		if enc1 == enc2 {
			t.Fatalf("EncryptString produced identical ciphertexts %q; nonce must be random", enc1)
		}

		dec, err := c.DecryptString(enc1)
		if err != nil {
			t.Fatalf("DecryptString(%q) err = %v", enc1, err)
		}
		if dec != plain {
			t.Fatalf("DecryptString() = %q, want %q", dec, plain)
		}
	}

	if crypt.IsEncryptedString("plain-text") {
		t.Fatalf("IsEncryptedString(\"plain-text\") = true, want false")
	}
}

func TestEncryptDecryptBytes(t *testing.T) {
	c, err := crypt.New(testSecret)
	if err != nil {
		t.Fatalf("New() err = %v", err)
	}

	ktpStub := bytes.Repeat([]byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F'}, 100)
	large := bytes.Repeat([]byte("A"), 1<<20)

	cases := [][]byte{
		{},
		ktpStub,
		large,
	}

	for _, plain := range cases {
		origCopy := append([]byte(nil), plain...)
		enc, err := c.EncryptBytes(plain)
		if err != nil {
			t.Fatalf("EncryptBytes(len=%d) err = %v", len(plain), err)
		}
		if !bytes.Equal(plain, origCopy) {
			t.Fatalf("EncryptBytes mutated caller plaintext slice")
		}
		if enc[0] != crypt.Version1 {
			t.Fatalf("wire version = %#x, want %#x", enc[0], crypt.Version1)
		}

		dec, err := c.DecryptBytes(enc)
		if err != nil {
			t.Fatalf("DecryptBytes(len=%d) err = %v", len(enc), err)
		}
		if !bytes.Equal(dec, plain) {
			t.Fatalf("DecryptBytes mismatch for len=%d", len(plain))
		}
	}
}

func TestTamperAndErrorCases(t *testing.T) {
	c1, err := crypt.New(testSecret)
	if err != nil {
		t.Fatalf("New(c1) err = %v", err)
	}
	c2, err := crypt.New("fedcba9876543210fedcba9876543210")
	if err != nil {
		t.Fatalf("New(c2) err = %v", err)
	}

	raw, err := c1.EncryptBytes([]byte("sensitive-ktp-blob"))
	if err != nil {
		t.Fatalf("EncryptBytes() err = %v", err)
	}

	// Unsupported version byte
	badVer := append([]byte(nil), raw...)
	badVer[0] = 0x02
	if _, err := c1.DecryptBytes(badVer); !errors.Is(err, crypt.ErrUnsupportedVersion) {
		t.Fatalf("DecryptBytes(badVer) err = %v, want %v", err, crypt.ErrUnsupportedVersion)
	}

	// Flip bit in nonce
	badNonce := append([]byte(nil), raw...)
	badNonce[2] ^= 0x01
	if _, err := c1.DecryptBytes(badNonce); !errors.Is(err, crypt.ErrDecryptFailed) {
		t.Fatalf("DecryptBytes(badNonce) err = %v, want %v", err, crypt.ErrDecryptFailed)
	}

	// Flip bit in ciphertext/tag
	badTag := append([]byte(nil), raw...)
	badTag[len(badTag)-1] ^= 0x01
	if _, err := c1.DecryptBytes(badTag); !errors.Is(err, crypt.ErrDecryptFailed) {
		t.Fatalf("DecryptBytes(badTag) err = %v, want %v", err, crypt.ErrDecryptFailed)
	}

	// Truncated payload
	if _, err := c1.DecryptBytes(raw[:10]); !errors.Is(err, crypt.ErrInvalidCiphertext) {
		t.Fatalf("DecryptBytes(truncated) err = %v, want %v", err, crypt.ErrInvalidCiphertext)
	}

	// Wrong secret
	if _, err := c2.DecryptBytes(raw); !errors.Is(err, crypt.ErrDecryptFailed) {
		t.Fatalf("DecryptBytes(wrong key) err = %v, want %v", err, crypt.ErrDecryptFailed)
	}

	// String errors: missing prefix, invalid base64, wrong key
	if _, err := c1.DecryptString("not-encrypted"); !errors.Is(err, crypt.ErrInvalidCiphertext) {
		t.Fatalf("DecryptString(missing prefix) err = %v, want %v", err, crypt.ErrInvalidCiphertext)
	}
	if _, err := c1.DecryptString(crypt.StringPrefixV1 + "!!!invalid-b64!!!"); !errors.Is(err, crypt.ErrInvalidCiphertext) {
		t.Fatalf("DecryptString(bad b64) err = %v, want %v", err, crypt.ErrInvalidCiphertext)
	}
	validStr := crypt.StringPrefixV1 + base64.RawURLEncoding.EncodeToString(raw)
	if _, err := c2.DecryptString(validStr); !errors.Is(err, crypt.ErrDecryptFailed) {
		t.Fatalf("DecryptString(wrong key) err = %v, want %v", err, crypt.ErrDecryptFailed)
	}
}

func TestBlindIndex(t *testing.T) {
	c1, err := crypt.New(testSecret)
	if err != nil {
		t.Fatalf("New(c1) err = %v", err)
	}
	c2, err := crypt.New("fedcba9876543210fedcba9876543210")
	if err != nil {
		t.Fatalf("New(c2) err = %v", err)
	}

	if got := c1.BlindIndex(""); got != "" {
		t.Fatalf("BlindIndex(\"\") = %q, want \"\"", got)
	}
	if got := c1.BlindIndex("   \t\n "); got != "" {
		t.Fatalf("BlindIndex(whitespace) = %q, want \"\"", got)
	}

	idx1 := c1.BlindIndex("3171010101900001")
	idx2 := c1.BlindIndex("  3171010101900001\n")
	if len(idx1) != 64 {
		t.Fatalf("len(BlindIndex) = %d, want 64", len(idx1))
	}
	if idx1 != idx2 {
		t.Fatalf("BlindIndex not deterministic across whitespace trim: %q != %q", idx1, idx2)
	}

	if idxOtherVal := c1.BlindIndex("3171010101900002"); idxOtherVal == idx1 {
		t.Fatalf("BlindIndex collision for different inputs")
	}
	if idxOtherKey := c2.BlindIndex("3171010101900001"); idxOtherKey == idx1 {
		t.Fatalf("BlindIndex should differ across different APP_SECRET keys")
	}
}

func BenchmarkEncryptBytes_1MB(b *testing.B) {
	c, err := crypt.New(testSecret)
	if err != nil {
		b.Fatal(err)
	}
	payload := bytes.Repeat([]byte("A"), 1<<20)
	b.SetBytes(int64(len(payload)))
	b.ResetTimer()
	for b.Loop() {
		if _, err := c.EncryptBytes(payload); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncryptString(b *testing.B) {
	c, err := crypt.New(testSecret)
	if err != nil {
		b.Fatal(err)
	}
	const secretVal = "sk-ant-api03-benchmark-key-1234567890"
	b.ResetTimer()
	for b.Loop() {
		if _, err := c.EncryptString(secretVal); err != nil {
			b.Fatal(err)
		}
	}
}

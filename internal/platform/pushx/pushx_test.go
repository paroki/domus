package pushx

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTestClientKeys(t *testing.T) (priv *ecdh.PrivateKey, p256dhB64, authB64 string) {
	t.Helper()
	var err error
	priv, err = ecdh.P256().GenerateKey(rand.Reader)
	require.NoError(t, err)
	auth := make([]byte, 16)
	_, err = io.ReadFull(rand.Reader, auth)
	require.NoError(t, err)
	return priv, base64.RawURLEncoding.EncodeToString(priv.PublicKey().Bytes()), base64.RawURLEncoding.EncodeToString(auth)
}

func decryptRFC8291(t *testing.T, uaPriv *ecdh.PrivateKey, authB64 string, encrypted []byte) []byte {
	t.Helper()
	require.GreaterOrEqual(t, len(encrypted), 21+65+16)
	salt := encrypted[:16]
	rs := binary.BigEndian.Uint32(encrypted[16:20])
	assert.Equal(t, defaultRecordSize, rs)
	idLen := int(encrypted[20])
	require.Equal(t, 65, idLen)
	asPubBytes := encrypted[21 : 21+idLen]
	ciphertext := encrypted[21+idLen:]

	asPub, err := ecdh.P256().NewPublicKey(asPubBytes)
	require.NoError(t, err)
	ecdhSecret, err := uaPriv.ECDH(asPub)
	require.NoError(t, err)

	authSecret, err := base64.RawURLEncoding.DecodeString(authB64)
	require.NoError(t, err)

	uaPubBytes := uaPriv.PublicKey().Bytes()
	keyInfo := append([]byte("WebPush: info\x00"), uaPubBytes...)
	keyInfo = append(keyInfo, asPubBytes...)

	ikm, err := hkdf.Key(sha256.New, ecdhSecret, authSecret, string(keyInfo), 32)
	require.NoError(t, err)
	cek, err := hkdf.Key(sha256.New, ikm, salt, "Content-Encoding: aes128gcm\x00", 16)
	require.NoError(t, err)
	nonce, err := hkdf.Key(sha256.New, ikm, salt, "Content-Encoding: nonce\x00", 12)
	require.NoError(t, err)

	block, err := aes.NewCipher(cek)
	require.NoError(t, err)
	gcm, err := cipher.NewGCM(block)
	require.NoError(t, err)

	padded, err := gcm.Open(nil, nonce, ciphertext, nil)
	require.NoError(t, err)
	require.NotEmpty(t, padded)
	require.Equal(t, byte(0x02), padded[len(padded)-1])
	return padded[:len(padded)-1]
}

func TestGenerateVAPIDKeys_AndValidateSubscription(t *testing.T) {
	pub, priv, err := GenerateVAPIDKeys()
	require.NoError(t, err)
	assert.NotEmpty(t, pub)
	assert.NotEmpty(t, priv)

	_, p256dh, auth := generateTestClientKeys(t)

	require.NoError(t, ValidateSubscription(Subscription{
		Endpoint: "https://fcm.googleapis.com/fcm/send/abc",
		P256DH:   p256dh,
		Auth:     auth,
	}))
	require.NoError(t, ValidateSubscription(Subscription{
		Endpoint: "http://127.0.0.1:9999/push",
		P256DH:   p256dh,
		Auth:     auth,
	}))

	// SSRF and invalid key guards
	for _, badEndpoint := range []string{
		"http://example.com/push",
		"https://10.0.0.1/push",
		"https://172.16.0.5/push",
		"https://192.168.1.10/push",
		"https://169.254.169.254/latest",
		"ftp://fcm.googleapis.com/push",
	} {
		err := ValidateSubscription(Subscription{Endpoint: badEndpoint, P256DH: p256dh, Auth: auth})
		assert.Error(t, err, "expected error for %s", badEndpoint)
	}

	assert.Error(t, ValidateSubscription(Subscription{
		Endpoint: "https://fcm.googleapis.com/fcm/send/abc",
		P256DH:   base64.RawURLEncoding.EncodeToString([]byte("short")),
		Auth:     auth,
	}))
	assert.Error(t, ValidateSubscription(Subscription{
		Endpoint: "https://fcm.googleapis.com/fcm/send/abc",
		P256DH:   p256dh,
		Auth:     base64.RawURLEncoding.EncodeToString([]byte("short")),
	}))
}

func TestSender_SendEncryptsPayloadAndSignsVAPID_AndDetectsExpired(t *testing.T) {
	pub, priv, err := GenerateVAPIDKeys()
	require.NoError(t, err)

	sender, err := New(Config{
		PublicKey:  pub,
		PrivateKey: priv,
		Subject:    "mailto:test@domus.org",
	})
	require.NoError(t, err)

	uaPriv, p256dh, auth := generateTestClientKeys(t)
	wantPlaintext := `{"title":"Domus","body":"Jadwal baru","url":"/events"}`

	var gotAuthHeader string
	var gotBody []byte
	var gotEncoding string
	statusCode := http.StatusCreated

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthHeader = r.Header.Get("Authorization")
		gotEncoding = r.Header.Get("Content-Encoding")
		var readErr error
		gotBody, readErr = io.ReadAll(r.Body)
		require.NoError(t, readErr)
		w.WriteHeader(statusCode)
	}))
	defer srv.Close()

	sub := Subscription{
		Endpoint: srv.URL + "/push/sub-1",
		P256DH:   p256dh,
		Auth:     auth,
	}

	require.NoError(t, sender.Send(context.Background(), sub, []byte(wantPlaintext)))
	assert.Equal(t, "aes128gcm", gotEncoding)
	assert.True(t, strings.HasPrefix(gotAuthHeader, "vapid t="))

	// Verify RFC 8291 decryption
	decrypted := decryptRFC8291(t, uaPriv, auth, gotBody)
	assert.Equal(t, wantPlaintext, string(decrypted))

	// Verify RFC 8292 ES256 JWT signature
	parts := strings.SplitN(strings.TrimPrefix(gotAuthHeader, "vapid t="), ", k=", 2)
	require.Len(t, parts, 2)
	assert.Equal(t, pub, parts[1])
	jwtParts := strings.Split(parts[0], ".")
	require.Len(t, jwtParts, 3)

	claimsJSON, err := base64.RawURLEncoding.DecodeString(jwtParts[1])
	require.NoError(t, err)
	var claims map[string]any
	require.NoError(t, json.Unmarshal(claimsJSON, &claims))
	assert.Equal(t, "mailto:test@domus.org", claims["sub"])
	assert.Equal(t, srv.URL, claims["aud"])

	pubBytes, err := base64.RawURLEncoding.DecodeString(pub)
	require.NoError(t, err)
	ecdsaPub, err := ecdsa.ParseUncompressedPublicKey(elliptic.P256(), pubBytes)
	require.NoError(t, err)
	sigBytes, err := base64.RawURLEncoding.DecodeString(jwtParts[2])
	require.NoError(t, err)
	require.Len(t, sigBytes, 64)
	digest := sha256.Sum256([]byte(jwtParts[0] + "." + jwtParts[1]))
	assert.True(t, ecdsa.Verify(ecdsaPub, digest[:], new(big.Int).SetBytes(sigBytes[:32]), new(big.Int).SetBytes(sigBytes[32:])))

	// 410 Gone -> ErrSubscriptionExpired
	statusCode = http.StatusGone
	err = sender.Send(context.Background(), sub, []byte(wantPlaintext))
	assert.ErrorIs(t, err, ErrSubscriptionExpired)

	// 404 Not Found -> ErrSubscriptionExpired
	statusCode = http.StatusNotFound
	err = sender.Send(context.Background(), sub, []byte(wantPlaintext))
	assert.ErrorIs(t, err, ErrSubscriptionExpired)
}

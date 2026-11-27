package pushx

import (
	"bytes"
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
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultRecordSize uint32 = 4096
	defaultTTLHeader         = "86400"
	vapidTokenTTL            = 12 * time.Hour
)

// ErrSubscriptionExpired indicates the push vendor rejected the subscription with 404 or 410.
var ErrSubscriptionExpired = errors.New("pushx: subscription expired")

// Subscription holds a browser Web Push subscription endpoint and client keys.
type Subscription struct {
	Endpoint string
	P256DH   string
	Auth     string
}

// Config configures a Web Push Sender with VAPID credentials.
type Config struct {
	PublicKey  string
	PrivateKey string
	Subject    string
	HTTPClient *http.Client
}

// Sender encrypts Web Push payloads (RFC 8291) and signs VAPID headers (RFC 8292).
type Sender struct {
	publicKeyB64 string
	privateKey   *ecdsa.PrivateKey
	subject      string
	httpClient   *http.Client
}

// GenerateVAPIDKeys generates an uncompressed P-256 public key and raw 32-byte private key encoded as base64url.
func GenerateVAPIDKeys() (publicKey, privateKey string, err error) {
	priv, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("pushx: generate vapid key: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(priv.PublicKey().Bytes()),
		base64.RawURLEncoding.EncodeToString(priv.Bytes()),
		nil
}

// New constructs a Sender from base64url-encoded VAPID keys and a contact subject.
func New(cfg Config) (*Sender, error) {
	pubRaw, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(cfg.PublicKey))
	if err != nil || len(pubRaw) != 65 || pubRaw[0] != 0x04 {
		return nil, errors.New("pushx: invalid vapid public key")
	}
	if _, err := ecdh.P256().NewPublicKey(pubRaw); err != nil {
		return nil, fmt.Errorf("pushx: invalid vapid public key: %w", err)
	}

	privRaw, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(cfg.PrivateKey))
	if err != nil || len(privRaw) != 32 {
		return nil, errors.New("pushx: invalid vapid private key")
	}
	if _, err := ecdh.P256().NewPrivateKey(privRaw); err != nil {
		return nil, fmt.Errorf("pushx: invalid vapid private key: %w", err)
	}

	ecdsaPriv, err := ecdsa.ParseRawPrivateKey(elliptic.P256(), privRaw)
	if err != nil {
		return nil, fmt.Errorf("pushx: invalid vapid private key: %w", err)
	}

	subject := strings.TrimSpace(cfg.Subject)
	if subject == "" {
		subject = "mailto:admin@domus.example.com"
	}

	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	return &Sender{
		publicKeyB64: base64.RawURLEncoding.EncodeToString(pubRaw),
		privateKey:   ecdsaPriv,
		subject:      subject,
		httpClient:   client,
	}, nil
}

// ValidateSubscription validates the endpoint URL (with SSRF guards) and client P-256/auth keys.
func ValidateSubscription(sub Subscription) error {
	endpoint := strings.TrimSpace(sub.Endpoint)
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return errors.New("pushx: invalid endpoint URL")
	}
	host := u.Hostname()
	isLocalhost := strings.EqualFold(host, "localhost") || host == "127.0.0.1" || host == "::1"
	switch strings.ToLower(u.Scheme) {
	case "https":
	case "http":
		if !isLocalhost {
			return errors.New("pushx: endpoint must use https")
		}
	default:
		return errors.New("pushx: endpoint must use https")
	}

	if ip := net.ParseIP(host); ip != nil && !isLocalhost {
		if ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast() {
			return errors.New("pushx: endpoint host IP is not allowed")
		}
	}

	if _, _, err := decodeSubscriptionKeys(sub); err != nil {
		return err
	}
	return nil
}

func decodeSubscriptionKeys(sub Subscription) (p256dh, auth []byte, err error) {
	p256dh, err = base64.RawURLEncoding.DecodeString(strings.TrimSpace(sub.P256DH))
	if err != nil || len(p256dh) != 65 || p256dh[0] != 0x04 {
		return nil, nil, errors.New("pushx: invalid p256dh key")
	}
	if _, err := ecdh.P256().NewPublicKey(p256dh); err != nil {
		return nil, nil, fmt.Errorf("pushx: invalid p256dh key: %w", err)
	}

	auth, err = base64.RawURLEncoding.DecodeString(strings.TrimSpace(sub.Auth))
	if err != nil || len(auth) != 16 {
		return nil, nil, errors.New("pushx: invalid auth secret")
	}
	return p256dh, auth, nil
}

// Encrypt encrypts plaintext using RFC 8291 (aes128gcm) for the given subscription.
func Encrypt(sub Subscription, plaintext []byte) ([]byte, error) {
	uaPubBytes, authSecret, err := decodeSubscriptionKeys(sub)
	if err != nil {
		return nil, err
	}
	uaPub, err := ecdh.P256().NewPublicKey(uaPubBytes)
	if err != nil {
		return nil, err
	}

	asPriv, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	asPub := asPriv.PublicKey().Bytes()

	ecdhSecret, err := asPriv.ECDH(uaPub)
	if err != nil {
		return nil, err
	}

	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	keyInfo := make([]byte, 0, len("WebPush: info\x00")+len(uaPubBytes)+len(asPub))
	keyInfo = append(keyInfo, "WebPush: info\x00"...)
	keyInfo = append(keyInfo, uaPubBytes...)
	keyInfo = append(keyInfo, asPub...)

	ikm, err := hkdf.Key(sha256.New, ecdhSecret, authSecret, string(keyInfo), 32)
	if err != nil {
		return nil, err
	}

	cek, err := hkdf.Key(sha256.New, ikm, salt, "Content-Encoding: aes128gcm\x00", 16)
	if err != nil {
		return nil, err
	}

	nonce, err := hkdf.Key(sha256.New, ikm, salt, "Content-Encoding: nonce\x00", 12)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(cek)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	record := make([]byte, len(plaintext)+1)
	copy(record, plaintext)
	record[len(plaintext)] = 0x02

	ciphertext := gcm.Seal(nil, nonce, record, nil)

	out := make([]byte, 0, 16+4+1+len(asPub)+len(ciphertext))
	out = append(out, salt...)
	var rsBuf [4]byte
	binary.BigEndian.PutUint32(rsBuf[:], defaultRecordSize)
	out = append(out, rsBuf[:]...)
	out = append(out, byte(len(asPub)))
	out = append(out, asPub...)
	out = append(out, ciphertext...)
	return out, nil
}

func (s *Sender) signVAPID(audience string, now time.Time) (string, error) {
	headerBytes, err := json.Marshal(map[string]string{
		"typ": "JWT",
		"alg": "ES256",
	})
	if err != nil {
		return "", err
	}
	claimsBytes, err := json.Marshal(map[string]any{
		"aud": audience,
		"exp": now.Add(vapidTokenTTL).Unix(),
		"sub": s.subject,
	})
	if err != nil {
		return "", err
	}

	signingInput := base64.RawURLEncoding.EncodeToString(headerBytes) + "." + base64.RawURLEncoding.EncodeToString(claimsBytes)
	digest := sha256.Sum256([]byte(signingInput))

	r, sv, err := ecdsa.Sign(rand.Reader, s.privateKey, digest[:])
	if err != nil {
		return "", err
	}
	sig := make([]byte, 64)
	r.FillBytes(sig[:32])
	sv.FillBytes(sig[32:])

	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

// Send encrypts payload and delivers it to sub.Endpoint with a signed VAPID Authorization header.
func (s *Sender) Send(ctx context.Context, sub Subscription, payload []byte) error {
	if err := ValidateSubscription(sub); err != nil {
		return err
	}
	body, err := Encrypt(sub, payload)
	if err != nil {
		return err
	}

	u, err := url.Parse(strings.TrimSpace(sub.Endpoint))
	if err != nil {
		return err
	}
	audience := u.Scheme + "://" + u.Host

	token, err := s.signVAPID(audience, time.Now().UTC())
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("vapid t=%s, k=%s", token, s.publicKeyB64))
	req.Header.Set("Content-Encoding", "aes128gcm")
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("TTL", defaultTTLHeader)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
		return ErrSubscriptionExpired
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("pushx: unexpected push status %d", resp.StatusCode)
	}
	return nil
}

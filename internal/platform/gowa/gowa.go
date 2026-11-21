package gowa

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	// SignatureHeader is the HTTP header carrying the GOWA HMAC-SHA256 webhook signature.
	SignatureHeader = "X-Hub-Signature-256"
	// DeviceIDHeader is the optional GOWA multi-device header.
	DeviceIDHeader = "X-Device-Id"
)

var e164Re = regexp.MustCompile(`^\+[1-9][0-9]{6,14}$`)

// Config configures a GOWA (go-whatsapp-web-multidevice) REST Client.
type Config struct {
	BaseURL       string
	BasicUser     string
	BasicPass     string
	DeviceID      string
	WebhookSecret string
	HTTPClient    *http.Client
}

// Client sends outbound WhatsApp messages via the GOWA REST API.
type Client struct {
	baseURL       string
	basicUser     string
	basicPass     string
	deviceID      string
	webhookSecret string
	httpClient    *http.Client
}

// New validates cfg and constructs a GOWA Client.
func New(cfg Config) (*Client, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		return nil, errors.New("gowa: base URL is required")
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, fmt.Errorf("gowa: invalid base URL %q", cfg.BaseURL)
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	return &Client{
		baseURL:       baseURL,
		basicUser:     strings.TrimSpace(cfg.BasicUser),
		basicPass:     cfg.BasicPass,
		deviceID:      strings.TrimSpace(cfg.DeviceID),
		webhookSecret: strings.TrimSpace(cfg.WebhookSecret),
		httpClient:    httpClient,
	}, nil
}

// VerifyWebhookSignature validates GOWA's X-Hub-Signature-256 header using this client's secret.
func (c *Client) VerifyWebhookSignature(body []byte, signatureHeader string) bool {
	if c == nil {
		return false
	}
	return VerifyWebhookSignature(c.webhookSecret, body, signatureHeader)
}

// FormatTargetJID resolves a personal phone/JID or WhatsApp group JID (@g.us) into a GOWA target JID.
func FormatTargetJID(target string) (string, error) {
	raw := strings.TrimSpace(target)
	if raw == "" {
		return "", errors.New("gowa: target phone or JID is required")
	}
	if at := strings.LastIndex(raw, "@"); at != -1 && strings.EqualFold(raw[at:], "@g.us") {
		local := strings.TrimSpace(raw[:at])
		if local == "" {
			return "", errors.New("gowa: invalid group JID")
		}
		return local + "@g.us", nil
	}
	e164, err := NormalizeJID(raw)
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(e164, "+") + "@s.whatsapp.net", nil
}

// SendMessage sends a WhatsApp text message to target (E.164, personal JID, or @g.us group JID) via GOWA POST /send/message.
func (c *Client) SendMessage(ctx context.Context, target, message string) error {
	if strings.TrimSpace(message) == "" {
		return errors.New("gowa: message text is required")
	}
	jid, err := FormatTargetJID(target)
	if err != nil {
		return err
	}
	return c.postJSON(ctx, "/send/message", map[string]string{
		"phone":   jid,
		"message": message,
	})
}

// SendChatPresence sends a transient chat presence indicator (e.g. "composing") via GOWA POST /send/chat-presence.
func (c *Client) SendChatPresence(ctx context.Context, target, action string) error {
	action = strings.TrimSpace(action)
	if action == "" {
		return errors.New("gowa: chat presence action is required")
	}
	jid, err := FormatTargetJID(target)
	if err != nil {
		return err
	}
	return c.postJSON(ctx, "/send/chat-presence", map[string]string{
		"phone":  jid,
		"action": action,
	})
}

func (c *Client) postJSON(ctx context.Context, path string, payload any) error {
	rawBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("gowa: marshal %s request: %w", path, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(rawBody))
	if err != nil {
		return fmt.Errorf("gowa: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.basicUser != "" || c.basicPass != "" {
		req.SetBasicAuth(c.basicUser, c.basicPass)
	}
	if c.deviceID != "" {
		req.Header.Set(DeviceIDHeader, c.deviceID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("gowa: %s request failed: %w", path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("gowa: %s returned HTTP %d: %s", path, resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

// VerifyWebhookSignature validates GOWA's X-Hub-Signature-256 ("sha256=<hex>") header in constant time.
// ponytail: APP_GOWA_WEBHOOK_SECRET env wiring deferred to Phase 6 MVP-6.4.
func VerifyWebhookSignature(secret string, body []byte, signatureHeader string) bool {
	if secret == "" {
		return false
	}
	const prefix = "sha256="
	if len(signatureHeader) <= len(prefix) || !strings.EqualFold(signatureHeader[:len(prefix)], prefix) {
		return false
	}
	gotMAC, err := hex.DecodeString(strings.TrimSpace(signatureHeader[len(prefix):]))
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(gotMAC, expectedMAC)
}

// NormalizeJID converts a personal WhatsApp JID or phone string into canonical E.164 (+628...).
// Non-personal JIDs (@g.us, @lid, @broadcast) are rejected.
func NormalizeJID(jid string) (string, error) {
	raw := strings.TrimSpace(jid)
	if raw == "" {
		return "", errors.New("gowa: phone or JID is required")
	}

	local := raw
	if at := strings.Index(raw, "@"); at != -1 {
		domain := strings.ToLower(raw[at:])
		switch domain {
		case "@s.whatsapp.net":
			local = raw[:at]
		case "@g.us", "@lid", "@broadcast":
			return "", fmt.Errorf("gowa: unsupported non-personal JID domain %q", domain)
		default:
			return "", fmt.Errorf("gowa: unsupported JID domain %q", domain)
		}
	}

	if colon := strings.Index(local, ":"); colon != -1 {
		local = local[:colon]
	}

	cleaned := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "").Replace(strings.TrimSpace(local))
	if cleaned == "" {
		return "", errors.New("gowa: invalid phone or JID")
	}
	if strings.HasPrefix(cleaned, "0") {
		cleaned = "+62" + cleaned[1:]
	} else if !strings.HasPrefix(cleaned, "+") {
		cleaned = "+" + cleaned
	}
	if !e164Re.MatchString(cleaned) {
		return "", fmt.Errorf("gowa: invalid E.164 phone %q", jid)
	}
	return cleaned, nil
}

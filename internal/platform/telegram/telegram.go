package telegram

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the default Telegram Bot API endpoint.
	DefaultBaseURL = "https://api.telegram.org"
	// SecretTokenHeader is the HTTP header carrying the Telegram webhook secret token.
	SecretTokenHeader = "X-Telegram-Bot-Api-Secret-Token"
)

// Config configures a Telegram Bot API Client.
type Config struct {
	BotToken      string
	BotUsername   string
	WebhookSecret string
	BaseURL       string
	HTTPClient    *http.Client
}

// Client sends outbound Telegram Bot API requests and verifies inbound webhook secrets.
type Client struct {
	botToken      string
	botUsername   string
	webhookSecret string
	baseURL       string
	httpClient    *http.Client
}

// Update represents an inbound Telegram Bot webhook update payload.
type Update struct {
	UpdateID int64    `json:"update_id"`
	Message  *Message `json:"message,omitempty"`
}

// Message represents a Telegram message inside an Update.
type Message struct {
	MessageID int64  `json:"message_id"`
	From      *User  `json:"from,omitempty"`
	Chat      Chat   `json:"chat"`
	Text      string `json:"text,omitempty"`
}

// Chat represents a Telegram chat destination.
type Chat struct {
	ID   int64  `json:"id"`
	Type string `json:"type,omitempty"`
}

// User represents a Telegram user sender.
type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// New validates cfg and constructs a Telegram Bot Client.
func New(cfg Config) (*Client, error) {
	token := strings.TrimSpace(cfg.BotToken)
	if token == "" {
		return nil, errors.New("telegram: bot token is required")
	}
	username := strings.TrimPrefix(strings.TrimSpace(cfg.BotUsername), "@")
	if username == "" {
		return nil, errors.New("telegram: bot username is required")
	}

	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, fmt.Errorf("telegram: invalid base URL %q", cfg.BaseURL)
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	return &Client{
		botToken:      token,
		botUsername:   username,
		webhookSecret: strings.TrimSpace(cfg.WebhookSecret),
		baseURL:       baseURL,
		httpClient:    httpClient,
	}, nil
}

// BotUsername returns the normalized Telegram bot username without leading '@'.
func (c *Client) BotUsername() string {
	return c.botUsername
}

// DeepLinkURL builds a Telegram deep link URL carrying start payload token.
func (c *Client) DeepLinkURL(token string) string {
	return fmt.Sprintf("https://t.me/%s?start=%s", c.botUsername, url.QueryEscape(strings.TrimSpace(token)))
}

// VerifyWebhookSecret verifies X-Telegram-Bot-Api-Secret-Token against the configured secret.
func (c *Client) VerifyWebhookSecret(headerValue string) bool {
	return VerifyWebhookSecret(c.webhookSecret, headerValue)
}

// SendMessage sends a plain-text message to chatID via the Telegram Bot API.
func (c *Client) SendMessage(ctx context.Context, chatID int64, text string) error {
	if chatID == 0 {
		return errors.New("telegram: chat id is required")
	}
	if strings.TrimSpace(text) == "" {
		return errors.New("telegram: message text is required")
	}
	return c.postJSON(ctx, "sendMessage", map[string]any{
		"chat_id": chatID,
		"text":    text,
	})
}

// SendChatAction sends a transient chat action (e.g. "typing") to chatID.
func (c *Client) SendChatAction(ctx context.Context, chatID int64, action string) error {
	if chatID == 0 {
		return errors.New("telegram: chat id is required")
	}
	action = strings.TrimSpace(action)
	if action == "" {
		return errors.New("telegram: chat action is required")
	}
	return c.postJSON(ctx, "sendChatAction", map[string]any{
		"chat_id": chatID,
		"action":  action,
	})
}

func (c *Client) postJSON(ctx context.Context, method string, payload any) error {
	rawBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("telegram: marshal %s request: %w", method, err)
	}

	endpoint := fmt.Sprintf("%s/bot%s/%s", c.baseURL, url.PathEscape(c.botToken), method)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(rawBody))
	if err != nil {
		return fmt.Errorf("telegram: build %s request: %w", method, c.redactErr(err))
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("telegram: %s request failed: %w", method, c.redactErr(err))
	}
	defer func() { _ = resp.Body.Close() }()

	respBytes, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return fmt.Errorf("telegram: read %s response: %w", method, err)
	}

	var apiResp struct {
		OK          bool   `json:"ok"`
		ErrorCode   int    `json:"error_code"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("telegram: %s returned HTTP %d", method, resp.StatusCode)
		}
		return fmt.Errorf("telegram: decode %s response: %w", method, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || !apiResp.OK {
		return fmt.Errorf("telegram: %s failed (status=%d code=%d): %s", method, resp.StatusCode, apiResp.ErrorCode, apiResp.Description)
	}
	return nil
}

func (c *Client) redactErr(err error) error {
	if err == nil {
		return nil
	}
	var ue *url.Error
	if errors.As(err, &ue) {
		safe := *ue
		safe.URL = strings.ReplaceAll(safe.URL, c.botToken, "[REDACTED]")
		return &safe
	}
	return errors.New(strings.ReplaceAll(err.Error(), c.botToken, "[REDACTED]"))
}

// VerifyWebhookSecret compares expectedSecret and headerValue in constant time (fail-closed when empty).
func VerifyWebhookSecret(expectedSecret, headerValue string) bool {
	if expectedSecret == "" || headerValue == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expectedSecret), []byte(headerValue)) == 1
}

// ParseStartToken extracts the deep-link token from a "/start <token>" or "/start@bot <token>" message.
func ParseStartToken(text string) (string, bool) {
	fields := strings.Fields(strings.TrimSpace(text))
	if len(fields) < 2 {
		return "", false
	}
	cmd := strings.ToLower(fields[0])
	if cmd != "/start" && (!strings.HasPrefix(cmd, "/start@") || len(cmd) <= len("/start@")) {
		return "", false
	}
	return fields[1], true
}

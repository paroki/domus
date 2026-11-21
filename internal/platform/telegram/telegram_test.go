package telegram_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paroki/domus/internal/platform/telegram"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidationAndDeepLink(t *testing.T) {
	_, err := telegram.New(telegram.Config{BotUsername: "domus_bot"})
	require.Error(t, err)

	_, err = telegram.New(telegram.Config{BotToken: "secret-token"})
	require.Error(t, err)

	_, err = telegram.New(telegram.Config{
		BotToken:    "secret-token",
		BotUsername: "domus_bot",
		BaseURL:     "ftp://invalid.example.com",
	})
	require.Error(t, err)

	c, err := telegram.New(telegram.Config{
		BotToken:      "secret-token",
		BotUsername:   "@domus_paroki_bot",
		WebhookSecret: "hook-secret-123",
	})
	require.NoError(t, err)
	assert.Equal(t, "domus_paroki_bot", c.BotUsername())
	assert.Equal(t, "https://t.me/domus_paroki_bot?start=abc123", c.DeepLinkURL("abc123"))
	assert.True(t, c.VerifyWebhookSecret("hook-secret-123"))
	assert.False(t, c.VerifyWebhookSecret("wrong-secret"))
}

func TestVerifyWebhookSecret(t *testing.T) {
	assert.False(t, telegram.VerifyWebhookSecret("", ""))
	assert.False(t, telegram.VerifyWebhookSecret("", "anything"))
	assert.False(t, telegram.VerifyWebhookSecret("expected", ""))
	assert.False(t, telegram.VerifyWebhookSecret("expected", "wrong"))
	assert.True(t, telegram.VerifyWebhookSecret("expected", "expected"))
}

func TestParseStartToken(t *testing.T) {
	cases := []struct {
		name      string
		text      string
		wantToken string
		wantOK    bool
	}{
		{name: "simple start", text: "/start tok-123", wantToken: "tok-123", wantOK: true},
		{name: "with bot mention", text: "  /start@domus_bot  tok_ABC  ", wantToken: "tok_ABC", wantOK: true},
		{name: "uppercase command", text: "/START tok-upper", wantToken: "tok-upper", wantOK: true},
		{name: "missing token", text: "/start", wantToken: "", wantOK: false},
		{name: "empty bot mention", text: "/start@ tok", wantToken: "", wantOK: false},
		{name: "different command", text: "/startnow tok", wantToken: "", wantOK: false},
		{name: "help command", text: "/help tok", wantToken: "", wantOK: false},
		{name: "plain text", text: "hello bot", wantToken: "", wantOK: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tok, ok := telegram.ParseStartToken(tc.text)
			assert.Equal(t, tc.wantOK, ok)
			assert.Equal(t, tc.wantToken, tok)
		})
	}
}

func TestSendMessageAndChatAction(t *testing.T) {
	var paths []string
	var payloads []map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		payloads = append(payloads, body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"result":{}}`))
	}))
	defer srv.Close()

	c, err := telegram.New(telegram.Config{
		BotToken:    "12345:SECRET_TOKEN",
		BotUsername: "domus_bot",
		BaseURL:     srv.URL,
	})
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, c.SendMessage(ctx, 987654321, "Kode OTP: 123456"))
	require.NoError(t, c.SendChatAction(ctx, 987654321, "typing"))

	require.Len(t, paths, 2)
	assert.Equal(t, "/bot12345:SECRET_TOKEN/sendMessage", paths[0])
	assert.Equal(t, float64(987654321), payloads[0]["chat_id"])
	assert.Equal(t, "Kode OTP: 123456", payloads[0]["text"])

	assert.Equal(t, "/bot12345:SECRET_TOKEN/sendChatAction", paths[1])
	assert.Equal(t, float64(987654321), payloads[1]["chat_id"])
	assert.Equal(t, "typing", payloads[1]["action"])
}

func TestSendMessage_ErrorsNeverLeakToken(t *testing.T) {
	const token = "99999:SUPER_SECRET_BOT_TOKEN"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"ok":false,"error_code":400,"description":"Bad Request: chat not found"}`))
	}))
	defer srv.Close()

	c, err := telegram.New(telegram.Config{
		BotToken:    token,
		BotUsername: "domus_bot",
		BaseURL:     srv.URL,
	})
	require.NoError(t, err)

	err = c.SendMessage(context.Background(), 123, "hello")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "chat not found")
	assert.NotContains(t, err.Error(), token)

	// Closed server triggers transport *url.Error which must also redact token.
	srv.Close()
	err = c.SendMessage(context.Background(), 123, "hello")
	require.Error(t, err)
	assert.NotContains(t, err.Error(), token)
}

func TestMessageHasMedia_AndEntities(t *testing.T) {
	var msg telegram.Message
	assert.False(t, msg.HasMedia())

	var nilMsg *telegram.Message
	assert.False(t, nilMsg.HasMedia())

	raw := `{
		"message_id": 10,
		"chat": {"id": -100123, "type": "supergroup"},
		"caption": "@domus_bot info",
		"caption_entities": [{"type": "mention", "offset": 0, "length": 10}],
		"photo": [{"file_id": "abc"}]
	}`
	require.NoError(t, json.Unmarshal([]byte(raw), &msg))
	assert.True(t, msg.HasMedia())
	require.Len(t, msg.CaptionEntities, 1)
	assert.Equal(t, "mention", msg.CaptionEntities[0].Type)
}

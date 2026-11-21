package gowa_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paroki/domus/internal/platform/gowa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidation(t *testing.T) {
	_, err := gowa.New(gowa.Config{})
	require.Error(t, err)

	_, err = gowa.New(gowa.Config{BaseURL: "ftp://wa.example.com"})
	require.Error(t, err)

	_, err = gowa.New(gowa.Config{BaseURL: "https://"})
	require.Error(t, err)

	c, err := gowa.New(gowa.Config{BaseURL: "https://gowa.example.com/"})
	require.NoError(t, err)
	require.NotNil(t, c)
}

func TestNormalizeJID(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "personal jid", input: "628123456789@s.whatsapp.net", want: "+628123456789"},
		{name: "multi-device jid", input: "628123456789:12@s.whatsapp.net", want: "+628123456789"},
		{name: "uppercase domain", input: "628123456789@S.WHATSAPP.NET", want: "+628123456789"},
		{name: "local zero prefix", input: "0812-3456-789", want: "+628123456789"},
		{name: "canonical e164", input: "+628123456789", want: "+628123456789"},
		{name: "group jid rejected", input: "120363000111222@g.us", wantErr: true},
		{name: "lid jid rejected", input: "123456789@lid", wantErr: true},
		{name: "broadcast jid rejected", input: "status@broadcast", wantErr: true},
		{name: "empty rejected", input: "   ", wantErr: true},
		{name: "non-numeric rejected", input: "abc@s.whatsapp.net", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := gowa.NormalizeJID(tc.input)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestVerifyWebhookSignature(t *testing.T) {
	secret := "gowa-secret-xyz"
	body := []byte(`{"event":"message","from":"628123456789@s.whatsapp.net"}`)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	assert.True(t, gowa.VerifyWebhookSignature(secret, body, sig))
	assert.False(t, gowa.VerifyWebhookSignature("", body, sig))
	assert.False(t, gowa.VerifyWebhookSignature(secret, body, ""))
	assert.False(t, gowa.VerifyWebhookSignature(secret, body, "md5=deadbeef"))
	assert.False(t, gowa.VerifyWebhookSignature(secret, []byte(`{"tampered":true}`), sig))
	assert.False(t, gowa.VerifyWebhookSignature("wrong-secret", body, sig))
}

func TestSendMessage(t *testing.T) {
	var gotPath, gotUser, gotPass, gotDevice string
	var gotOK bool
	var gotBody map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotUser, gotPass, gotOK = r.BasicAuth()
		gotDevice = r.Header.Get(gowa.DeviceIDHeader)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":"SUCCESS"}`))
	}))
	defer srv.Close()

	c, err := gowa.New(gowa.Config{
		BaseURL:   srv.URL,
		BasicUser: "admin",
		BasicPass: "secret-pass",
		DeviceID:  "device-01",
	})
	require.NoError(t, err)

	require.NoError(t, c.SendMessage(context.Background(), "+628123456789", "Kode OTP: 654321"))
	assert.Equal(t, "/send/message", gotPath)
	assert.True(t, gotOK)
	assert.Equal(t, "admin", gotUser)
	assert.Equal(t, "secret-pass", gotPass)
	assert.Equal(t, "device-01", gotDevice)
	assert.Equal(t, "628123456789@s.whatsapp.net", gotBody["phone"])
	assert.Equal(t, "Kode OTP: 654321", gotBody["message"])
}

func TestSendMessage_UpstreamFailureDoesNotLeakPass(t *testing.T) {
	const pass = "super-secret-gowa-pass"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
	}))
	defer srv.Close()

	c, err := gowa.New(gowa.Config{
		BaseURL:   srv.URL,
		BasicUser: "admin",
		BasicPass: pass,
	})
	require.NoError(t, err)

	err = c.SendMessage(context.Background(), "+628123456789", "hello")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
	assert.NotContains(t, err.Error(), pass)
}

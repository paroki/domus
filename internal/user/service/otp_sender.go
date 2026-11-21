package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/paroki/domus/internal/platform/gowa"
	"github.com/paroki/domus/internal/platform/telegram"
	"github.com/paroki/domus/internal/user/domain"
)

const channelWhatsAppLink = "whatsapp_link"

// GowaOTPSender delivers WhatsApp OTP codes via GOWA REST API.
type GowaOTPSender struct {
	client *gowa.Client
	otpTTL time.Duration
}

// NewGowaOTPSender constructs a GowaOTPSender backed by client.
func NewGowaOTPSender(client *gowa.Client, otpTTL time.Duration) *GowaOTPSender {
	return &GowaOTPSender{client: client, otpTTL: otpTTL}
}

// SendOTP sends a 6-digit OTP message to phone target over WhatsApp.
func (s *GowaOTPSender) SendOTP(ctx context.Context, channel, target, code string) error {
	if s.client == nil {
		return errors.New("gowa otp sender: client not configured")
	}
	ttlMins := max(1, int(s.otpTTL.Minutes()))
	var msg string
	switch channel {
	case domain.ProviderWhatsApp:
		msg = fmt.Sprintf(
			"Kode OTP login Domus Anda adalah: %s\n\nKode ini berlaku selama %d menit. Jangan berikan kode ini kepada siapa pun.",
			code,
			ttlMins,
		)
	case channelWhatsAppLink:
		msg = fmt.Sprintf(
			"Kode verifikasi WhatsApp Domus Anda adalah: %s\n\nKode ini berlaku selama %d menit. Jangan berikan kode ini kepada siapa pun.",
			code,
			ttlMins,
		)
	default:
		return fmt.Errorf("gowa otp sender: unsupported channel %q", channel)
	}
	return s.client.SendMessage(ctx, target, msg)
}

// TelegramOTPSender delivers Telegram OTP codes via Telegram Bot API.
type TelegramOTPSender struct {
	client *telegram.Client
	otpTTL time.Duration
}

// NewTelegramOTPSender constructs a TelegramOTPSender backed by client.
func NewTelegramOTPSender(client *telegram.Client, otpTTL time.Duration) *TelegramOTPSender {
	return &TelegramOTPSender{client: client, otpTTL: otpTTL}
}

// SendOTP sends a 6-digit OTP message to chatID target over Telegram.
func (s *TelegramOTPSender) SendOTP(ctx context.Context, channel, target, code string) error {
	if s.client == nil {
		return errors.New("telegram otp sender: client not configured")
	}
	if channel != "" && channel != domain.ProviderTelegram {
		return fmt.Errorf("telegram otp sender: unsupported channel %q", channel)
	}
	chatID, err := strconv.ParseInt(strings.TrimSpace(target), 10, 64)
	if err != nil || chatID == 0 {
		return fmt.Errorf("telegram otp sender: invalid chat id %q", target)
	}
	ttlMins := max(1, int(s.otpTTL.Minutes()))
	msg := fmt.Sprintf(
		"Kode OTP login Domus Anda adalah: %s\n\nKode ini berlaku selama %d menit. Jangan berikan kode ini kepada siapa pun.",
		code,
		ttlMins,
	)
	return s.client.SendMessage(ctx, chatID, msg)
}

// CompositeOTPSender routes OTP dispatches across configured delivery channels.
type CompositeOTPSender struct {
	mu      sync.RWMutex
	senders map[string]domain.OTPSender
}

// NewCompositeOTPSender creates a CompositeOTPSender with LogOTPSender as the default email sender.
func NewCompositeOTPSender(logger *slog.Logger) *CompositeOTPSender {
	return &CompositeOTPSender{
		senders: map[string]domain.OTPSender{
			domain.ProviderEmail: NewLogOTPSender(logger),
		},
	}
}

// SetSender registers or overrides the OTPSender for channel.
func (c *CompositeOTPSender) SetSender(channel string, sender domain.OTPSender) {
	ch := strings.ToLower(strings.TrimSpace(channel))
	if ch == "" || sender == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.senders[ch] = sender
}

// RemoveSender unregisters the OTPSender for channel.
func (c *CompositeOTPSender) RemoveSender(channel string) {
	ch := strings.ToLower(strings.TrimSpace(channel))
	if ch == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.senders, ch)
}

// HasChannel reports whether channel has a configured sender.
func (c *CompositeOTPSender) HasChannel(channel string) bool {
	ch := strings.ToLower(strings.TrimSpace(channel))
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.senders[ch] != nil
}

// EnabledChannels returns configured login OTP channels in deterministic order.
func (c *CompositeOTPSender) EnabledChannels() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var channels []string
	for _, ch := range []string{domain.ProviderEmail, domain.ProviderWhatsApp, domain.ProviderTelegram} {
		if c.senders[ch] != nil {
			channels = append(channels, ch)
		}
	}
	return channels
}

// SendOTP routes OTP delivery to the configured sender for channel (or whatsapp_link -> whatsapp).
func (c *CompositeOTPSender) SendOTP(ctx context.Context, channel, target, code string) error {
	ch := strings.ToLower(strings.TrimSpace(channel))
	if ch == "" {
		ch = domain.ProviderEmail
	}
	routeKey := ch
	if routeKey == channelWhatsAppLink {
		routeKey = domain.ProviderWhatsApp
	}
	c.mu.RLock()
	sender := c.senders[routeKey]
	c.mu.RUnlock()
	if sender == nil {
		return domain.ErrAuthChannelDisabled
	}
	return sender.SendOTP(ctx, ch, target, code)
}

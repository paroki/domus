package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/gowa"
	"github.com/paroki/domus/internal/platform/pushx"
	"github.com/paroki/domus/internal/platform/telegram"
	"github.com/paroki/domus/internal/user/domain"
)

const (
	ChannelPush     = "push"
	ChannelWhatsApp = "whatsapp"
	ChannelTelegram = "telegram"
)

// PushNotification delivers Web Push notifications via *pushx.Sender.
type PushNotification struct {
	sender *pushx.Sender
}

// NewPushNotification constructs a PushNotification backed by sender.
func NewPushNotification(sender *pushx.Sender) *PushNotification {
	return &PushNotification{sender: sender}
}

// SendNotification encrypts and sends payload to a JSON-encoded pushx.Subscription target.
func (n *PushNotification) SendNotification(ctx context.Context, target string, payload domain.NotificationPayload) error {
	if n.sender == nil {
		return errors.New("push notification: sender not configured")
	}
	var sub pushx.Subscription
	if err := json.Unmarshal([]byte(target), &sub); err != nil {
		return fmt.Errorf("push notification: invalid subscription target: %w", err)
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return n.sender.Send(ctx, sub, body)
}

// GowaNotification delivers WhatsApp notifications via *gowa.Client.
type GowaNotification struct {
	client *gowa.Client
}

// NewGowaNotification constructs a GowaNotification backed by client.
func NewGowaNotification(client *gowa.Client) *GowaNotification {
	return &GowaNotification{client: client}
}

// SendNotification sends payload as a formatted text message to phone target over WhatsApp.
func (n *GowaNotification) SendNotification(ctx context.Context, target string, payload domain.NotificationPayload) error {
	if n.client == nil {
		return errors.New("gowa notification: client not configured")
	}
	msg := formatTextNotification(payload)
	return n.client.SendMessage(ctx, target, msg)
}

// TelegramNotification delivers Telegram notifications via *telegram.Client.
type TelegramNotification struct {
	client *telegram.Client
}

// NewTelegramNotification constructs a TelegramNotification backed by client.
func NewTelegramNotification(client *telegram.Client) *TelegramNotification {
	return &TelegramNotification{client: client}
}

// SendNotification sends payload as a formatted text message to numeric chatID target over Telegram.
func (n *TelegramNotification) SendNotification(ctx context.Context, target string, payload domain.NotificationPayload) error {
	if n.client == nil {
		return errors.New("telegram notification: client not configured")
	}
	chatID, err := strconv.ParseInt(strings.TrimSpace(target), 10, 64)
	if err != nil || chatID == 0 {
		return fmt.Errorf("telegram notification: invalid chat id %q", target)
	}
	msg := formatTextNotification(payload)
	return n.client.SendMessage(ctx, chatID, msg)
}

func formatTextNotification(payload domain.NotificationPayload) string {
	var b strings.Builder
	if payload.Title != "" {
		b.WriteString(payload.Title)
	}
	if payload.Body != "" {
		if b.Len() > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString(payload.Body)
	}
	if payload.URL != "" {
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString(payload.URL)
	}
	return b.String()
}

// CompositeNotification is a thread-safe registry routing notifications across push, whatsapp, and telegram.
type CompositeNotification struct {
	mu       sync.RWMutex
	channels map[string]domain.NotificationSender
}

// NewCompositeNotification constructs an empty CompositeNotification registry.
func NewCompositeNotification() *CompositeNotification {
	return &CompositeNotification{
		channels: make(map[string]domain.NotificationSender),
	}
}

// SetChannel registers or overrides a NotificationSender for channel.
func (c *CompositeNotification) SetChannel(channel string, sender domain.NotificationSender) {
	ch := strings.ToLower(strings.TrimSpace(channel))
	if ch == "" || sender == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.channels[ch] = sender
}

// RemoveChannel unregisters the NotificationSender for channel.
func (c *CompositeNotification) RemoveChannel(channel string) {
	ch := strings.ToLower(strings.TrimSpace(channel))
	if ch == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.channels, ch)
}

// GetChannel returns the registered NotificationSender for channel if present.
func (c *CompositeNotification) GetChannel(channel string) (domain.NotificationSender, bool) {
	ch := strings.ToLower(strings.TrimSpace(channel))
	c.mu.RLock()
	defer c.mu.RUnlock()
	s, ok := c.channels[ch]
	return s, ok
}

// SendNotification dispatches payload to target via the named channel.
func (c *CompositeNotification) SendNotification(ctx context.Context, channel, target string, payload domain.NotificationPayload) error {
	s, ok := c.GetChannel(channel)
	if !ok || s == nil {
		return fmt.Errorf("notification: channel %q is not configured", channel)
	}
	return s.SendNotification(ctx, target, payload)
}

// GetVAPIDPublicKey returns the active VAPID public key for browser PushManager.subscribe.
func (s *Service) GetVAPIDPublicKey() (domain.VAPIDPublicKeyResponse, error) {
	s.vapidMu.RLock()
	pub := s.vapidPublicKey
	s.vapidMu.RUnlock()
	if pub == "" {
		return domain.VAPIDPublicKeyResponse{}, domain.ErrPushNotConfigured
	}
	return domain.VAPIDPublicKeyResponse{PublicKey: pub}, nil
}

// SubscribePush validates and upserts a browser Web Push subscription for the authenticated user.
func (s *Service) SubscribePush(ctx context.Context, p authx.Principal, userAgent string, req domain.SubscribePushRequest) error {
	sub := pushx.Subscription{
		Endpoint: strings.TrimSpace(req.Endpoint),
		P256DH:   strings.TrimSpace(req.Keys.P256DH),
		Auth:     strings.TrimSpace(req.Keys.Auth),
	}
	if err := pushx.ValidateSubscription(sub); err != nil {
		msg := err.Error()
		switch {
		case strings.Contains(msg, "p256dh"):
			return domain.ValidationError("keys.p256dh", "INVALID_PUSH_KEY", "Invalid p256dh public key.")
		case strings.Contains(msg, "auth"):
			return domain.ValidationError("keys.auth", "INVALID_PUSH_KEY", "Invalid auth secret.")
		default:
			return domain.ValidationError("endpoint", "INVALID_PUSH_ENDPOINT", "Push endpoint must be a valid public HTTPS URL.")
		}
	}

	var sessionID *uuid.UUID
	if p.SessionID != uuid.Nil {
		sid := p.SessionID
		sessionID = &sid
	}
	_, err := s.repo.UpsertPushSubscription(ctx, p.UserID, sessionID, sub.Endpoint, sub.P256DH, sub.Auth, userAgent)
	return err
}

// UnsubscribePush removes the matching Web Push subscription for the authenticated user.
func (s *Service) UnsubscribePush(ctx context.Context, p authx.Principal, req domain.UnsubscribePushRequest) error {
	endpoint := strings.TrimSpace(req.Endpoint)
	if endpoint == "" {
		return domain.ValidationError("endpoint", "REQUIRED", "Endpoint is required.")
	}
	_, err := s.repo.DeletePushSubscriptionByEndpoint(ctx, p.UserID, endpoint)
	return err
}

// TestPushNotification sends a test notification across the caller's enabled channels.
func (s *Service) TestPushNotification(ctx context.Context, p authx.Principal) error {
	u, err := guardActiveUser(s.repo.FindUserByID(ctx, p.UserID))
	if err != nil {
		return err
	}

	payload := domain.NotificationPayload{
		Title: "Domus",
		Body:  "Notifikasi uji coba berhasil dikirim.",
		Icon:  "/logo-192.png",
		Badge: "/logo-192.png",
		Tag:   "domus-test",
		URL:   "/me",
	}

	var firstErr error
	if u.NotifyPush {
		if err := s.DispatchPushToUser(ctx, u.ID, payload); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if u.NotifyWhatsapp && u.Phone != nil && *u.Phone != "" {
		if ch, ok := s.getNotificationChannel(ChannelWhatsApp); ok {
			if err := ch.SendNotification(ctx, *u.Phone, payload); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	if u.NotifyTelegram {
		if acc, accErr := s.repo.FindUserAccountByUserAndProvider(ctx, u.ID, domain.ProviderTelegram); accErr == nil && acc != nil {
			if ch, ok := s.getNotificationChannel(ChannelTelegram); ok {
				if err := ch.SendNotification(ctx, acc.Subject, payload); err != nil && firstErr == nil {
					firstErr = err
				}
			}
		}
	}
	return firstErr
}

// DispatchPushToUser delivers payload to all Web Push subscriptions of userID, pruning expired rows.
func (s *Service) DispatchPushToUser(ctx context.Context, userID uuid.UUID, payload domain.NotificationPayload) error {
	sender, ok := s.getNotificationChannel(ChannelPush)
	if !ok {
		return domain.ErrPushNotConfigured
	}
	subs, err := s.repo.ListPushSubscriptionsByUserID(ctx, userID)
	if err != nil {
		return err
	}
	var firstErr error
	for _, sub := range subs {
		targetBytes, err := json.Marshal(pushx.Subscription{
			Endpoint: sub.Endpoint,
			P256DH:   sub.P256dh,
			Auth:     sub.Auth,
		})
		if err != nil {
			continue
		}
		if sendErr := sender.SendNotification(ctx, string(targetBytes), payload); sendErr != nil {
			if errors.Is(sendErr, pushx.ErrSubscriptionExpired) {
				_ = s.repo.DeletePushSubscriptionByID(ctx, sub.ID)
				continue
			}
			if firstErr == nil {
				firstErr = sendErr
			}
			continue
		}
		_ = s.repo.TouchPushSubscriptionLastUsed(ctx, sub.ID, time.Now().UTC())
	}
	return firstErr
}

func (s *Service) getNotificationChannel(channel string) (domain.NotificationSender, bool) {
	if s.compositeNotification == nil {
		return nil, false
	}
	return s.compositeNotification.GetChannel(channel)
}

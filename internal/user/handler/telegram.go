package handler

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/platform/telegram"
	"github.com/paroki/domus/internal/user/domain"
)

// TelegramWebhook handles inbound Telegram Bot API updates.
//
//	@Summary		Telegram bot webhook
//	@Description	Receives Telegram Bot API updates and processes /start account linking tokens.
//	@ID				telegramWebhook
//	@Tags			Telegram
//	@Accept			json
//	@Produce		json
//	@Param			X-Telegram-Bot-Api-Secret-Token	header		string	false	"Optional Telegram webhook secret token"
//	@Param			body							body		object	true	"Telegram update payload"
//	@Success		200								{object}	httpx.Envelope[any]
//	@Failure		400								{object}	httpx.Envelope[any]
//	@Failure		401								{object}	httpx.Envelope[any]
//	@Router			/telegram/hooks [post]
func (h *Handler) TelegramWebhook(c fiber.Ctx) error {
	tgClient := h.getTelegramClient()
	if tgClient == nil {
		return httpx.Unauthorized("Telegram bot is not configured.")
	}
	if tgClient.HasWebhookSecret() && !tgClient.VerifyWebhookSecret(c.Get(telegram.SecretTokenHeader)) {
		return httpx.Unauthorized("Invalid Telegram webhook secret.")
	}

	var update telegram.Update
	if err := json.Unmarshal(c.Body(), &update); err != nil {
		return httpx.BadRequest("Invalid Telegram update payload.")
	}

	if update.Message != nil {
		chatID := update.Message.Chat.ID
		if token, ok := telegram.ParseStartToken(update.Message.Text); ok {
			if chatID != 0 {
				displayName := telegramDisplayName(update.Message.From)
				err := h.service.CompleteTelegramLink(c.Context(), token, strconv.FormatInt(chatID, 10), displayName)
				reply := telegramLinkReply(err)
				_ = tgClient.SendMessage(c.Context(), chatID, reply)
			}
		} else if h.telegramFallback != nil {
			_ = h.telegramFallback(c.Context(), update)
		} else if chatID != 0 {
			_ = tgClient.SendMessage(
				c.Context(),
				chatID,
				"Halo! Untuk menghubungkan akun Telegram dengan Domus, silakan buka halaman Profil (/me) di Domus dan buat tautan verifikasi Telegram.",
			)
		}
	}

	return httpx.OK(c, map[string]bool{"ok": true})
}

func telegramLinkReply(err error) string {
	switch {
	case err == nil:
		return "✅ Telegram berhasil terhubung ke akun Domus Anda!"
	case errors.Is(err, domain.ErrUserAccountConflict):
		return "⚠️ Akun Telegram ini sudah terhubung ke akun Domus lain."
	case errors.Is(err, domain.ErrUserDisabled):
		return "⚠️ Akun Domus Anda sedang dinonaktifkan."
	default:
		return "⚠️ Tautan verifikasi Telegram tidak valid atau sudah kedaluwarsa. Silakan buat tautan baru di Domus."
	}
}

func telegramDisplayName(from *telegram.User) string {
	if from == nil {
		return ""
	}
	full := strings.TrimSpace(strings.TrimSpace(from.FirstName) + " " + strings.TrimSpace(from.LastName))
	if full != "" {
		return full
	}
	if username := strings.TrimPrefix(strings.TrimSpace(from.Username), "@"); username != "" {
		return "@" + username
	}
	return ""
}

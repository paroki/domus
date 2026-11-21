package handler

import (
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/agent/service"
	"github.com/paroki/domus/internal/platform/gowa"
	"github.com/paroki/domus/internal/platform/httpx"
)

type gowaWebhookPayload struct {
	Event         string          `json:"event"`
	From          string          `json:"from"`
	ChatID        string          `json:"chat_id"`
	SenderID      string          `json:"sender_id"`
	Participant   string          `json:"participant"`
	FromMe        bool            `json:"from_me"`
	IsFromMe      bool            `json:"is_from_me"`
	Text          string          `json:"text"`
	Body          string          `json:"body"`
	Caption       string          `json:"caption"`
	MentionedJIDs []string        `json:"mentioned_jids"`
	Mentions      []string        `json:"mentions"`
	MediaType     string          `json:"media_type"`
	Image         any             `json:"image"`
	Audio         any             `json:"audio"`
	Video         any             `json:"video"`
	Sticker       any             `json:"sticker"`
	Document      any             `json:"document"`
	Message       *gowaMessageObj `json:"message"`
	Payload       json.RawMessage `json:"payload"`
	Data          json.RawMessage `json:"data"`
}

type gowaMessageObj struct {
	Text     string `json:"text"`
	Body     string `json:"body"`
	Caption  string `json:"caption"`
	Image    any    `json:"image"`
	Audio    any    `json:"audio"`
	Video    any    `json:"video"`
	Sticker  any    `json:"sticker"`
	Document any    `json:"document"`
}

// WhatsAppWebhook handles inbound GOWA WhatsApp webhook events for the AI agent.
//
//	@Summary		GOWA WhatsApp agent webhook
//	@Description	Verifies X-Hub-Signature-256 and routes inbound WhatsApp messages to the AI agent.
//	@ID				whatsAppAgentWebhook
//	@Tags			Agent
//	@Accept			json
//	@Produce		json
//	@Param			X-Hub-Signature-256	header		string	true	"GOWA HMAC-SHA256 signature"
//	@Param			body				body		object	true	"GOWA webhook payload"
//	@Success		200					{object}	httpx.Envelope[any]
//	@Failure		400					{object}	httpx.Envelope[any]
//	@Failure		401					{object}	httpx.Envelope[any]
//	@Router			/agent/webhooks/whatsapp [post]
func (h *Handler) WhatsAppWebhook(c fiber.Ctx) error {
	gowaClient := h.service.GowaClient()
	if gowaClient == nil || !gowaClient.VerifyWebhookSignature(c.Body(), c.Get(gowa.SignatureHeader)) {
		return domain.ErrWebhookUnauthorized
	}

	var top gowaWebhookPayload
	if err := json.Unmarshal(c.Body(), &top); err != nil {
		return httpx.BadRequest("Invalid WhatsApp webhook payload.")
	}

	inbound := normalizeGowaPayload(top)
	if len(top.Payload) > 0 && string(top.Payload) != "null" {
		var nested gowaWebhookPayload
		if err := json.Unmarshal(top.Payload, &nested); err == nil {
			if nested.Event == "" {
				nested.Event = top.Event
			}
			inbound = normalizeGowaPayload(nested)
		}
	} else if len(top.Data) > 0 && string(top.Data) != "null" {
		var nested gowaWebhookPayload
		if err := json.Unmarshal(top.Data, &nested); err == nil {
			if nested.Event == "" {
				nested.Event = top.Event
			}
			inbound = normalizeGowaPayload(nested)
		}
	}

	_ = h.service.HandleWhatsAppInbound(c.Context(), inbound)
	return httpx.OK(c, map[string]bool{"ok": true})
}

func normalizeGowaPayload(p gowaWebhookPayload) service.WhatsAppInbound {
	text := firstString(p.Text, p.Body)
	caption := p.Caption
	hasMedia := strings.TrimSpace(p.MediaType) != "" ||
		p.Image != nil || p.Audio != nil || p.Video != nil || p.Sticker != nil || p.Document != nil

	if p.Message != nil {
		text = firstString(text, p.Message.Text, p.Message.Body)
		caption = firstString(caption, p.Message.Caption)
		if p.Message.Image != nil || p.Message.Audio != nil || p.Message.Video != nil || p.Message.Sticker != nil || p.Message.Document != nil {
			hasMedia = true
		}
	}
	mentions := p.MentionedJIDs
	if len(mentions) == 0 {
		mentions = p.Mentions
	}
	return service.WhatsAppInbound{
		Event:         p.Event,
		From:          p.From,
		ChatID:        p.ChatID,
		SenderID:      p.SenderID,
		Participant:   p.Participant,
		FromMe:        p.FromMe || p.IsFromMe,
		Text:          text,
		Caption:       caption,
		MentionedJIDs: mentions,
		HasMedia:      hasMedia,
	}
}

func firstString(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

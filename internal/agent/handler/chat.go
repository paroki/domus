package handler

import (
	"bufio"
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/agent/service"
	"github.com/paroki/domus/internal/platform/authx"
	_ "github.com/paroki/domus/internal/platform/httpx"
)

// Chat streams a multi-agent AI assistant response over Server-Sent Events (SSE).
//
//	@Summary		Chat with AI assistant (SSE)
//	@Description	Validates request, resolves active model/session, and streams SSE events (session, agent_transfer, tool_start, tool_end, token, done, error).
//	@ID				chatWithAgent
//	@Tags			Agent
//	@Accept			json
//	@Produce		text/event-stream
//	@Param			body	body		domain.ChatRequest	true	"Chat request"
//	@Success		200		{string}	string				"text/event-stream"
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/agent/chat [post]
func (h *Handler) Chat(c fiber.Ctx) error {
	p, err := authx.RequireTenant(c.Context())
	if err != nil {
		return err
	}
	var req domain.ChatRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	prep, err := h.chat.PrepareChat(c.Context(), p, req)
	if err != nil {
		return err
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	return c.SendStreamWriter(func(w *bufio.Writer) {
		h.chat.ExecuteChatStream(prep, func(ev service.StreamEvent) error {
			if ev.Event == "ping" {
				if _, err := w.WriteString(": ping\n\n"); err != nil {
					return err
				}
				return w.Flush()
			}
			data, err := json.Marshal(ev.Data)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Event, data); err != nil {
				return err
			}
			return w.Flush()
		})
	})
}

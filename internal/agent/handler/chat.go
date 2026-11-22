package handler

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/agent/service"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/errx"
	"github.com/paroki/domus/internal/platform/httpx"
)

// Chat executes an AI agent conversation turn and streams named Server-Sent Events.
//
//	@Summary		Chat with AI agent (SSE)
//	@Description	Streams a multi-agent ReAct conversation turn over Server-Sent Events. Pre-stream validation and session lock errors return JSON envelopes.
//	@ID				chatWithAgent
//	@Tags			Agent
//	@Accept			json
//	@Produce		text/event-stream
//	@Param			body	body		domain.AgentChatRequest	true	"Chat request"
//	@Success		200		{string}	string					"Named SSE stream (session, thinking, tool_call, tool_result, delta, done, error)"
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		503		{object}	httpx.Envelope[any]
//	@Router			/agent/chat [post]
func (h *Handler) Chat(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.AgentChatRequest
	if err := decodeJSONOnly(c, &req); err != nil {
		return err
	}
	prep, err := h.service.PrepareTurn(c.Context(), service.ChatTurnInput{
		Principal: p,
		Channel:   domain.ChannelWeb,
		SessionID: req.SessionID,
		Message:   req.Message,
	})
	if err != nil {
		return err
	}

	c.Set(fiber.HeaderContentType, "text/event-stream")
	c.Set(fiber.HeaderCacheControl, "no-cache")
	c.Set(fiber.HeaderConnection, "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	// Capture an independent context because fasthttp recycles fiber.Ctx after the handler returns.
	// Cancel on client disconnect (write error) or after chatTurnTimeout.
	streamCtx, cancel := context.WithTimeout(context.Background(), chatTurnTimeout)
	streamCtx = authx.WithPrincipal(streamCtx, p)
	c.RequestCtx().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer cancel()
		defer prep.Release()
		_, _, runErr := h.service.RunPreparedTurn(streamCtx, prep, func(ev domain.AgentSSEEvent) error {
			if err := writeSSEFrame(w, ev); err != nil {
				cancel()
				return err
			}
			return nil
		})
		if runErr != nil {
			_ = writeSSEFrame(w, errorSSEEvent(runErr))
		}
	})
	return nil
}

// chatTurnTimeout bounds a single SSE chat turn.
const chatTurnTimeout = 5 * time.Minute

func writeSSEFrame(w *bufio.Writer, ev domain.AgentSSEEvent) error {
	b, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Event, b); err != nil {
		return err
	}
	return w.Flush()
}

func errorSSEEvent(err error) domain.AgentSSEEvent {
	code := "INTERNAL_ERROR"
	msg := err.Error()
	var ex *errx.Error
	if errors.As(err, &ex) {
		if ex.Code != "" {
			code = ex.Code
		}
		if ex.Message != "" {
			msg = ex.Message
		}
	}
	return domain.AgentSSEEvent{
		Event:   "error",
		Code:    code,
		Message: msg,
	}
}

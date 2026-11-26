package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/internal/umat/service"
)

// Handler exposes HTTP endpoints for the umat feature.
type Handler struct {
	service *service.Service
}

// New constructs a umat Handler backed by svc.
func New(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

// Register mounts /umat routes onto api (/api).
func (h *Handler) Register(_ fiber.Router) {}

package handler

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/config"
	"github.com/paroki/domus/api/internal/delivery/gofiber/response"
)

var Version = "dev"
var startTime = time.Now()

type HealthHandler struct {
	cfg *config.Config
}

func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{cfg: cfg}
}

type HealthResponse struct {
	Status  string  `json:"status"`
	Version string  `json:"version"`
	Env     string  `json:"env"`
	Uptime  float64 `json:"uptime"`
}

func (h *HealthHandler) Check(c fiber.Ctx) error {
	uptimeSeconds := time.Since(startTime).Seconds()

	data := HealthResponse{
		Status:  "pass",
		Version: Version,
		Env:     h.cfg.Env,
		Uptime:  uptimeSeconds,
	}

	return response.OK(c, data)
}

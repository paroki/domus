package app

import (
	"errors"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/internal/platform/httpx"
)

// InitSentry configures the global Sentry client. With an empty DSN the SDK is
// a no-op. The returned func flushes buffered events and must be deferred.
func InitSentry(cfg Config) (func(), error) {
	// ponytail: fixed server name & 0.1 trace sample rate; add to Config when per-env tuning needed.
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.SentryDSN,
		Environment:      cfg.SentryEnvironment,
		Release:          cfg.SentryRelease,
		ServerName:       "domus",
		EnableTracing:    true,
		TracesSampleRate: 0.1,
	})
	return func() { sentry.Flush(2 * time.Second) }, err
}

// sentryMiddleware reports panics (re-panicking for recover) and 5xx errors.
// ponytail: sentry-go/fiber targets Fiber v2; replace with it once v3 lands.
// Tracing spans per request are not recorded yet.
func sentryMiddleware(c fiber.Ctx) error {
	hub := sentry.CurrentHub().Clone()
	hub.Scope().SetTag("method", c.Method())
	hub.Scope().SetTag("path", c.Path())
	defer func() {
		if r := recover(); r != nil {
			hub.Recover(r)
			panic(r)
		}
	}()
	err := c.Next()
	if err != nil && errStatus(err) >= fiber.StatusInternalServerError {
		hub.CaptureException(err)
	}
	return err
}

func errStatus(err error) int {
	var ae *httpx.AppError
	if errors.As(err, &ae) {
		return ae.Status
	}
	var fe *fiber.Error
	if errors.As(err, &fe) {
		return fe.Code
	}
	return fiber.StatusInternalServerError
}

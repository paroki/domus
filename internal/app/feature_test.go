package app

import (
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

type stubFeature struct {
	prefixes []string
}

func (s stubFeature) Register(api fiber.Router) {
	for _, p := range s.prefixes {
		prefix := p
		api.Group(prefix).Get("/ping", func(c fiber.Ctx) error {
			return c.SendString(prefix)
		})
	}
}

func TestMountFeatures(t *testing.T) {
	app := fiber.New()
	mountFeatures(
		app.Group("/api"),
		stubFeature{prefixes: []string{"/auth", "/users"}},
		stubFeature{prefixes: []string{"/parishioners"}},
	)

	for _, prefix := range []string{"/auth", "/users", "/parishioners"} {
		resp, err := app.Test(httptest.NewRequest("GET", "/api"+prefix+"/ping", nil))
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != 200 || string(body) != prefix {
			t.Fatalf("GET /api%s/ping: got %d %q", prefix, resp.StatusCode, body)
		}
	}
}

type proxyProbe struct {
	BaseURL string `json:"baseUrl"`
	Scheme  string `json:"scheme"`
	IP      string `json:"ip"`
}

type proxyProbeFeature struct{}

func (proxyProbeFeature) Register(api fiber.Router) {
	api.Get("/proxy-probe", func(c fiber.Ctx) error {
		return c.JSON(proxyProbe{
			BaseURL: c.BaseURL(),
			Scheme:  c.Scheme(),
			IP:      c.IP(),
		})
	})
}

func TestTrustProxyHeaders(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)

	a := GetFiber(Config{}, logger, proxyProbeFeature{})
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- a.Listener(ln, fiber.ListenConfig{DisableStartupMessage: true})
	}()
	t.Cleanup(func() {
		_ = a.Shutdown()
		if err := <-errCh; err != nil {
			t.Errorf("listener error: %v", err)
		}
	})

	req, err := http.NewRequest(http.MethodGet, "http://"+ln.Addr().String()+"/api/proxy-probe", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "domus.example.com")
	req.Header.Set("X-Forwarded-For", "203.0.113.10")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	var trusted proxyProbe
	if err := json.NewDecoder(resp.Body).Decode(&trusted); err != nil {
		t.Fatal(err)
	}
	if trusted.Scheme != "https" || trusted.BaseURL != "https://domus.example.com" || trusted.IP != "203.0.113.10" {
		t.Fatalf("TrustProxy: got %+v", trusted)
	}
}

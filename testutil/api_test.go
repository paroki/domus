package testutil_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/testutil"
	"github.com/paroki/domus/testutil/mocks"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -source=api_test.go -destination=mocks/mock_ping_service.go -package=mocks

// PingService is a consumer-defined interface used to verify mockgen + ApiTestSuite.
type PingService interface {
	Ping(ctx context.Context, msg string) (string, error)
	List(ctx context.Context, page, limit int) ([]string, int, error)
	Upload(ctx context.Context, filename string, size int) (string, error)
}

type pingPayload struct {
	Message string `json:"message"`
}

type pingResponse struct {
	Reply   string `json:"reply"`
	Header  string `json:"header,omitempty"`
	Cookie  string `json:"cookie,omitempty"`
	Auth    string `json:"auth,omitempty"`
	TraceID string `json:"trace_id,omitempty"`
}

type pingFeature struct {
	svc PingService
}

func (f pingFeature) Register(api fiber.Router) {
	g := api.Group("/pings")

	g.Post("", func(c fiber.Ctx) error {
		var req pingPayload
		if err := c.Bind().Body(&req); err != nil {
			return err
		}
		reply, err := f.svc.Ping(c.Context(), req.Message)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(httpx.Envelope[any]{
				Error: &httpx.ErrorPayload{
					Code:      "PING_FAILED",
					Message:   err.Error(),
					RequestID: "req-123",
					Details:   []httpx.FieldError{{Field: "message", Code: "INVALID", Message: "invalid"}},
				},
			})
		}
		return c.Status(fiber.StatusCreated).JSON(httpx.Envelope[pingResponse]{
			Data: pingResponse{
				Reply:  reply,
				Header: c.Get("X-Custom"),
				Cookie: c.Cookies("session"),
				Auth:   c.Get("Authorization"),
			},
		})
	})

	g.Get("", func(c fiber.Ctx) error {
		items, total, err := f.svc.List(c.Context(), 1, 2)
		if err != nil {
			return err
		}
		return c.Status(fiber.StatusOK).JSON(httpx.Envelope[[]pingResponse]{
			Data: []pingResponse{{Reply: items[0]}, {Reply: items[1]}},
			Meta: &httpx.Meta{NextCursor: "cursor-2", HasMore: true, Limit: 2, Total: &total},
		})
	})

	g.Post("/upload", func(c fiber.Ctx) error {
		fh, err := c.FormFile("file")
		if err != nil {
			return err
		}
		file, err := fh.Open()
		if err != nil {
			return err
		}
		defer func() { _ = file.Close() }()
		raw, err := io.ReadAll(file)
		if err != nil {
			return err
		}
		stored, err := f.svc.Upload(c.Context(), fh.Filename, len(raw))
		if err != nil {
			return err
		}
		return c.Status(fiber.StatusOK).JSON(httpx.Envelope[pingResponse]{
			Data: pingResponse{Reply: stored},
		})
	})

	g.Get("/status/:code", func(c fiber.Ctx) error {
		code := fiber.Params[int](c, "code")
		if code == fiber.StatusNoContent {
			return c.SendStatus(fiber.StatusNoContent)
		}
		return c.Status(code).JSON(httpx.Envelope[any]{
			Error: &httpx.ErrorPayload{
				Code:    "STATUS_ERROR",
				Message: "status error",
			},
		})
	})
}

type PingApiSuite struct {
	testutil.ApiTestSuite[pingResponse]
	ctrl *gomock.Controller
	svc  *mocks.MockPingService
}

func (s *PingApiSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	s.ctrl = gomock.NewController(s.T())
	s.svc = mocks.NewMockPingService(s.ctrl)
	s.WithFeatures(pingFeature{svc: s.svc})
}

func (s *PingApiSuite) TestHealthDefaultApp() {
	var base testutil.ApiTestSuite[string]
	base.SetT(s.T())
	base.SetupSuite()
	base.SetupTest()
	base.Request("/api/health", http.MethodGet, nil)
	base.OK()
}

func (s *PingApiSuite) TestCreateAndHeadersCookiesToken() {
	s.svc.EXPECT().
		Ping(gomock.Any(), "hello").
		Return("pong:hello", nil)

	s.SetHeader("X-Custom", "parish-1")
	s.SetCookie(&http.Cookie{Name: "session", Value: "abc123"})
	s.RequestWithToken("/api/pings", http.MethodPost, pingPayload{Message: "hello"}, "tok-1")

	s.Created()
	env := s.GetResponse()
	s.Equal("pong:hello", env.Data.Reply)
	s.Equal("parish-1", env.Data.Header)
	s.Equal("abc123", env.Data.Cookie)
	s.Equal("Bearer tok-1", env.Data.Auth)

	// Verify RequestWithToken did not mutate s.BearerToken, and ClearCookies works.
	s.ClearCookies()
	s.svc.EXPECT().
		Ping(gomock.Any(), "second").
		Return("pong:second", nil)
	s.Request("/api/pings", http.MethodPost, pingPayload{Message: "second"})
	s.Created()
	second := testutil.DecodeAs[pingResponse](s)
	s.Empty(second.Data.Cookie)
	s.Empty(second.Data.Auth)
}

func (s *PingApiSuite) TestPagedList() {
	s.svc.EXPECT().
		List(gomock.Any(), 1, 2).
		Return([]string{"a", "b"}, 3, nil)

	s.Request("/api/pings", http.MethodGet, nil)
	s.OK()

	env := s.PagedResponse()
	s.Len(env.Data, 2)
	s.Equal("a", env.Data[0].Reply)
	s.Require().NotNil(env.Meta)
	s.Equal("cursor-2", env.Meta.NextCursor)
	s.True(env.Meta.HasMore)
	s.Equal(2, env.Meta.Limit)
	s.Require().NotNil(env.Meta.Total)
	s.Equal(3, *env.Meta.Total)
}

func (s *PingApiSuite) TestMultipartUpload() {
	s.svc.EXPECT().
		Upload(gomock.Any(), "note.txt", 11).
		Return("stored:note.txt", nil)

	s.RequestMultipart("/api/pings/upload", http.MethodPost, "file", "note.txt", []byte("hello world"))
	s.OK()

	env := s.GetResponse()
	s.Equal("stored:note.txt", env.Data.Reply)
}

func (s *PingApiSuite) TestErrorAndStatusHelpers() {
	s.svc.EXPECT().
		Ping(gomock.Any(), "bad").
		Return("", fiber.NewError(fiber.StatusBadRequest, "bad ping"))

	s.Request("/api/pings", http.MethodPost, pingPayload{Message: "bad"})
	s.BadRequest()
	errEnv := s.GetErrorResponse()
	s.Require().NotNil(errEnv.Error)
	s.Equal("PING_FAILED", errEnv.Error.Code)
	s.Equal("req-123", errEnv.Error.RequestID)
	s.Require().Len(errEnv.Error.Details, 1)
	s.Equal("INVALID", errEnv.Error.Details[0].Code)

	s.Request("/api/pings/status/204", http.MethodGet, nil)
	s.NoContent()

	s.Request("/api/pings/status/401", http.MethodGet, nil)
	s.Unauthorized()

	s.Request("/api/pings/status/403", http.MethodGet, nil)
	s.Forbidden()

	s.Request("/api/pings/status/404", http.MethodGet, nil)
	s.NotFound()

	s.Request("/api/pings/status/409", http.MethodGet, nil)
	s.Conflict()

	s.Request("/api/pings/status/422", http.MethodGet, nil)
	s.UnprocessableEntity()
}

func TestPingApiSuite(t *testing.T) {
	suite.Run(t, new(PingApiSuite))
}

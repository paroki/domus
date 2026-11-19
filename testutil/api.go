package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/internal/app"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Envelope mirrors the standard Domus API response shape { data, meta, error }.
type Envelope[T any] struct {
	Data  T             `json:"data"`
	Meta  *Meta         `json:"meta,omitempty"`
	Error *ErrorPayload `json:"error,omitempty"`
}

// Meta holds cursor pagination metadata returned in list responses.
type Meta struct {
	NextCursor string `json:"nextCursor,omitempty"`
	HasMore    bool   `json:"hasMore"`
	Limit      int    `json:"limit,omitempty"`
	Total      *int   `json:"total,omitempty"`
}

// ErrorPayload represents the structured error object in Envelope.Error.
type ErrorPayload struct {
	Code      string       `json:"code"`
	Message   string       `json:"message"`
	RequestID string       `json:"requestId,omitempty"`
	Details   []FieldError `json:"details,omitempty"`
}

// FieldError represents a single field-level validation error.
type FieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

// ApiTestSuite provides reusable Fiber v3 HTTP test helpers for feature suites.
type ApiTestSuite[T any] struct {
	suite.Suite
	App          *fiber.App
	HttpResponse *http.Response
	Cookies      []*http.Cookie
	Headers      http.Header
	BearerToken  string
	Context      context.Context
}

// WithFeatures bootstraps s.App with the given features and a discard logger.
func (s *ApiTestSuite[T]) WithFeatures(features ...app.Feature) *ApiTestSuite[T] {
	s.App = app.GetFiber(app.Config{}, slog.New(slog.DiscardHandler), features...)
	return s
}

// SetApp assigns a preconfigured Fiber app to the suite.
func (s *ApiTestSuite[T]) SetApp(a *fiber.App) *ApiTestSuite[T] {
	s.App = a
	return s
}

// SetupSuite initializes s.App with defaults if not already configured.
func (s *ApiTestSuite[T]) SetupSuite() {
	if s.App == nil {
		s.App = app.GetFiber(app.Config{}, slog.New(slog.DiscardHandler))
	}
}

// SetupTest resets per-test request and response state.
func (s *ApiTestSuite[T]) SetupTest() {
	s.Cookies = nil
	s.Headers = make(http.Header)
	s.BearerToken = ""
	s.HttpResponse = nil
	s.Context = context.Background()
}

// GetHttpResponse returns the last recorded HTTP response.
func (s *ApiTestSuite[T]) GetHttpResponse() *http.Response {
	return s.HttpResponse
}

// SetHeader sets a request header for subsequent requests in the current test.
func (s *ApiTestSuite[T]) SetHeader(key, value string) {
	if s.Headers == nil {
		s.Headers = make(http.Header)
	}
	s.Headers.Set(key, value)
}

// SetCookie appends a cookie for subsequent requests in the current test.
func (s *ApiTestSuite[T]) SetCookie(cookie *http.Cookie) {
	if cookie != nil {
		s.Cookies = append(s.Cookies, cookie)
	}
}

// ClearCookies removes all queued request cookies for the current test.
func (s *ApiTestSuite[T]) ClearCookies() {
	s.Cookies = nil
}

// Request executes an HTTP request against s.App and records s.HttpResponse.
func (s *ApiTestSuite[T]) Request(path, method string, requestBody any) {
	s.T().Helper()
	if s.App == nil {
		s.SetupSuite()
	}

	var bodyReader io.Reader
	var contentType string

	if requestBody != nil {
		switch v := requestBody.(type) {
		case io.Reader:
			bodyReader = v
		case []byte:
			bodyReader = bytes.NewReader(v)
		case string:
			bodyReader = strings.NewReader(v)
		default:
			payload, err := json.Marshal(v)
			s.Require().NoError(err, "marshal request body")
			bodyReader = bytes.NewReader(payload)
			contentType = "application/json"
		}
	}

	req := httptest.NewRequestWithContext(s.requestContext(), method, path, bodyReader)
	s.applyRequestMeta(req, contentType)

	resp, err := s.App.Test(req)
	s.Require().NoError(err, "execute fiber test request")
	s.HttpResponse = resp
}

// RequestWithToken executes Request with a temporary Bearer token without mutating s.BearerToken.
func (s *ApiTestSuite[T]) RequestWithToken(path, method string, requestBody any, token string) {
	s.T().Helper()
	prev := s.BearerToken
	s.BearerToken = token
	defer func() { s.BearerToken = prev }()
	s.Request(path, method, requestBody)
}

// RequestMultipart executes a multipart/form-data request containing a single file field.
func (s *ApiTestSuite[T]) RequestMultipart(path, method, fieldName, filename string, content []byte) {
	s.T().Helper()
	if s.App == nil {
		s.SetupSuite()
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile(fieldName, filename)
	s.Require().NoError(err, "create multipart form file")
	_, err = part.Write(content)
	s.Require().NoError(err, "write multipart content")
	s.Require().NoError(writer.Close(), "close multipart writer")

	req := httptest.NewRequestWithContext(s.requestContext(), method, path, &buf)
	s.applyRequestMeta(req, "")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := s.App.Test(req)
	s.Require().NoError(err, "execute fiber multipart request")
	s.HttpResponse = resp
}

func (s *ApiTestSuite[T]) requestContext() context.Context {
	if s.Context != nil {
		return s.Context
	}
	return context.Background()
}

func (s *ApiTestSuite[T]) applyRequestMeta(req *http.Request, defaultContentType string) {
	for k, vals := range s.Headers {
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}
	if defaultContentType != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", defaultContentType)
	}
	if s.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.BearerToken)
	}
	for _, c := range s.Cookies {
		req.AddCookie(c)
	}
}

// GetResponse decodes s.HttpResponse.Body into Envelope[T].
func (s *ApiTestSuite[T]) GetResponse() Envelope[T] {
	s.T().Helper()
	return DecodeAs[T](s)
}

// PagedResponse decodes s.HttpResponse.Body into Envelope[[]T].
func (s *ApiTestSuite[T]) PagedResponse() Envelope[[]T] {
	s.T().Helper()
	return DecodeAs[[]T](s)
}

// GetErrorResponse decodes s.HttpResponse.Body into Envelope[any].
func (s *ApiTestSuite[T]) GetErrorResponse() Envelope[any] {
	s.T().Helper()
	return DecodeAs[any](s)
}

// DecodeAs decodes the JSON response body from s.GetHttpResponse() into Envelope[R].
func DecodeAs[R any](s interface{ GetHttpResponse() *http.Response }) Envelope[R] {
	var t *testing.T
	if tp, ok := s.(interface{ T() *testing.T }); ok {
		t = tp.T()
		if t != nil {
			t.Helper()
		}
	}

	resp := s.GetHttpResponse()
	if t != nil {
		require.NotNil(t, resp, "http response is nil")
	} else if resp == nil {
		panic("testutil.DecodeAs: http response is nil")
	}

	raw, err := io.ReadAll(resp.Body)
	if t != nil {
		require.NoError(t, err, "read response body")
	} else if err != nil {
		panic(err)
	}
	_ = resp.Body.Close()
	// Restore body so subsequent reads/assertions on the same response still work.
	resp.Body = io.NopCloser(bytes.NewReader(raw))

	var env Envelope[R]
	err = json.Unmarshal(raw, &env)
	if t != nil {
		require.NoError(t, err, "unmarshal response envelope: %s", string(raw))
	} else if err != nil {
		panic(err)
	}
	return env
}

// AssertStatus asserts that s.HttpResponse has the expected HTTP status code.
func (s *ApiTestSuite[T]) AssertStatus(code int) {
	s.T().Helper()
	s.Require().NotNil(s.HttpResponse, "http response is nil")
	s.Assert().Equal(code, s.HttpResponse.StatusCode)
}

// OK asserts HTTP 200.
func (s *ApiTestSuite[T]) OK() {
	s.T().Helper()
	s.AssertStatus(http.StatusOK)
}

// Created asserts HTTP 201.
func (s *ApiTestSuite[T]) Created() {
	s.T().Helper()
	s.AssertStatus(http.StatusCreated)
}

// NoContent asserts HTTP 204.
func (s *ApiTestSuite[T]) NoContent() {
	s.T().Helper()
	s.AssertStatus(http.StatusNoContent)
}

// BadRequest asserts HTTP 400.
func (s *ApiTestSuite[T]) BadRequest() {
	s.T().Helper()
	s.AssertStatus(http.StatusBadRequest)
}

// Unauthorized asserts HTTP 401.
func (s *ApiTestSuite[T]) Unauthorized() {
	s.T().Helper()
	s.AssertStatus(http.StatusUnauthorized)
}

// Forbidden asserts HTTP 403.
func (s *ApiTestSuite[T]) Forbidden() {
	s.T().Helper()
	s.AssertStatus(http.StatusForbidden)
}

// NotFound asserts HTTP 404.
func (s *ApiTestSuite[T]) NotFound() {
	s.T().Helper()
	s.AssertStatus(http.StatusNotFound)
}

// Conflict asserts HTTP 409.
func (s *ApiTestSuite[T]) Conflict() {
	s.T().Helper()
	s.AssertStatus(http.StatusConflict)
}

// UnprocessableEntity asserts HTTP 422.
func (s *ApiTestSuite[T]) UnprocessableEntity() {
	s.T().Helper()
	s.AssertStatus(http.StatusUnprocessableEntity)
}

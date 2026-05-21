package testutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

// RequestBuilder builds an HTTP request and executes it against a Fiber app.
type RequestBuilder struct {
	app     *fiber.App
	method  string
	path    string
	body    io.Reader
	headers map[string]string
	err     error
}

// New creates a new RequestBuilder instance.
func New(app *fiber.App) *RequestBuilder {
	return &RequestBuilder{
		app:     app,
		headers: make(map[string]string),
	}
}

// GET sets the request method to GET and the request path.
func (b *RequestBuilder) GET(path string) *RequestBuilder {
	b.method = "GET"
	b.path = path
	return b
}

// POST sets the request method to POST and the request path.
func (b *RequestBuilder) POST(path string) *RequestBuilder {
	b.method = "POST"
	b.path = path
	return b
}

// PUT sets the request method to PUT and the request path.
func (b *RequestBuilder) PUT(path string) *RequestBuilder {
	b.method = "PUT"
	b.path = path
	return b
}

// PATCH sets the request method to PATCH and the request path.
func (b *RequestBuilder) PATCH(path string) *RequestBuilder {
	b.method = "PATCH"
	b.path = path
	return b
}

// DELETE sets the request method to DELETE and the request path.
func (b *RequestBuilder) DELETE(path string) *RequestBuilder {
	b.method = "DELETE"
	b.path = path
	return b
}

// OPTIONS sets the request method to OPTIONS and the request path.
func (b *RequestBuilder) OPTIONS(path string) *RequestBuilder {
	b.method = "OPTIONS"
	b.path = path
	return b
}

// Method sets an arbitrary request method and path.
func (b *RequestBuilder) Method(method, path string) *RequestBuilder {
	b.method = method
	b.path = path
	return b
}

// WithJSON marshals the provided value as JSON and sets it as the request body.
// It also sets the Content-Type header to application/json.
func (b *RequestBuilder) WithJSON(v any) *RequestBuilder {
	data, err := json.Marshal(v)
	if err != nil {
		b.err = fmt.Errorf("marshal JSON: %w", err)
		return b
	}
	b.body = bytes.NewReader(data)
	b.headers["Content-Type"] = "application/json"
	return b
}

// WithBody sets the request body and Content-Type header.
func (b *RequestBuilder) WithBody(r io.Reader, contentType string) *RequestBuilder {
	b.body = r
	b.headers["Content-Type"] = contentType
	return b
}

// WithHeader sets an HTTP header on the request.
func (b *RequestBuilder) WithHeader(key, value string) *RequestBuilder {
	b.headers[key] = value
	return b
}

// WithBearerToken sets the Authorization header with a Bearer token.
func (b *RequestBuilder) WithBearerToken(token string) *RequestBuilder {
	b.headers["Authorization"] = "Bearer " + token
	return b
}

// AssertionBuilder performs assertions on the HTTP response.
type AssertionBuilder struct {
	t          *testing.T
	resp       *http.Response
	body       []byte
	jsonParsed any
	failed     bool
}

// Expect executes the request and returns an AssertionBuilder.
func (b *RequestBuilder) Expect(t *testing.T) *AssertionBuilder {
	if b.err != nil {
		t.Fatalf("request builder setup error: %v", b.err)
	}

	req, err := http.NewRequest(b.method, b.path, b.body)
	if err != nil {
		t.Fatalf("failed to create http request: %v", err)
	}

	for k, v := range b.headers {
		req.Header.Set(k, v)
	}

	resp, err := b.app.Test(req)
	if err != nil {
		t.Fatalf("app test run failed: %v", err)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	var jsonParsed any
	_ = json.Unmarshal(bodyBytes, &jsonParsed)

	return &AssertionBuilder{
		t:          t,
		resp:       resp,
		body:       bodyBytes,
		jsonParsed: jsonParsed,
	}
}

// logFailure logs the response details on the first failure.
func (a *AssertionBuilder) logFailure(format string, args ...any) {
	if !a.failed {
		a.failed = true
		a.t.Logf("--- FAILED TEST HTTP RESPONSE ---")
		a.t.Logf("Status:  %s", a.resp.Status)
		a.t.Logf("Headers: %v", a.resp.Header)
		a.t.Logf("Body:    %s", string(a.body))
		a.t.Logf("---------------------------------")
	}
}

// Status asserts that the response status code matches the expected value.
// Fails the test immediately if there is a mismatch.
func (a *AssertionBuilder) Status(code int) *AssertionBuilder {
	if a.resp.StatusCode != code {
		a.logFailure("expected status code %d, got %d", code, a.resp.StatusCode)
		a.t.Fatalf("expected status code %d, got %d", code, a.resp.StatusCode)
	}
	return a
}

// StatusOK asserts that the response status code is 200.
func (a *AssertionBuilder) StatusOK() *AssertionBuilder {
	return a.Status(http.StatusOK)
}

// StatusCreated asserts that the response status code is 201.
func (a *AssertionBuilder) StatusCreated() *AssertionBuilder {
	return a.Status(http.StatusCreated)
}

// StatusNoContent asserts that the response status code is 204.
func (a *AssertionBuilder) StatusNoContent() *AssertionBuilder {
	return a.Status(http.StatusNoContent)
}

// StatusBadRequest asserts that the response status code is 400.
func (a *AssertionBuilder) StatusBadRequest() *AssertionBuilder {
	return a.Status(http.StatusBadRequest)
}

// StatusUnauthorized asserts that the response status code is 401.
func (a *AssertionBuilder) StatusUnauthorized() *AssertionBuilder {
	return a.Status(http.StatusUnauthorized)
}

// StatusForbidden asserts that the response status code is 403.
func (a *AssertionBuilder) StatusForbidden() *AssertionBuilder {
	return a.Status(http.StatusForbidden)
}

// StatusNotFound asserts that the response status code is 404.
func (a *AssertionBuilder) StatusNotFound() *AssertionBuilder {
	return a.Status(http.StatusNotFound)
}

// HeaderAssertionBuilder performs assertions on response headers.
type HeaderAssertionBuilder struct {
	assertion *AssertionBuilder
	name      string
	values    []string
}

// Header returns a HeaderAssertionBuilder for the specified header name.
func (a *AssertionBuilder) Header(name string) *HeaderAssertionBuilder {
	return &HeaderAssertionBuilder{
		assertion: a,
		name:      name,
		values:    a.resp.Header[name],
	}
}

// Eq asserts that the header value equals the expected string.
func (h *HeaderAssertionBuilder) Eq(want string) *AssertionBuilder {
	got := strings.Join(h.values, ", ")
	if got != want {
		h.assertion.logFailure("expected header %q to be %q, got %q", h.name, want, got)
		h.assertion.t.Errorf("expected header %q to be %q, got %q", h.name, want, got)
	}
	return h.assertion
}

// Exists asserts that the header exists in the response.
func (h *HeaderAssertionBuilder) Exists() *AssertionBuilder {
	if len(h.values) == 0 {
		h.assertion.logFailure("expected header %q to exist, but it was missing", h.name)
		h.assertion.t.Errorf("expected header %q to exist, but it was missing", h.name)
	}
	return h.assertion
}

// DoesNotExist asserts that the header does not exist in the response.
func (h *HeaderAssertionBuilder) DoesNotExist() *AssertionBuilder {
	if len(h.values) > 0 {
		h.assertion.logFailure("expected header %q to not exist, but got %q", h.name, strings.Join(h.values, ", "))
		h.assertion.t.Errorf("expected header %q to not exist, but got %q", h.name, strings.Join(h.values, ", "))
	}
	return h.assertion
}

// JSONAssertionBuilder performs assertions on parsed JSON values.
type JSONAssertionBuilder struct {
	assertion *AssertionBuilder
	path      string
	value     any
	exists    bool
	err       error
}

// JSONPath evaluates the specified dot-notation path on the JSON response.
func (a *AssertionBuilder) JSONPath(path string) *JSONAssertionBuilder {
	val, exists, err := evaluatePath(a.jsonParsed, path)
	return &JSONAssertionBuilder{
		assertion: a,
		path:      path,
		value:     val,
		exists:    exists,
		err:       err,
	}
}

// Exists asserts that the JSON path exists.
func (j *JSONAssertionBuilder) Exists() *AssertionBuilder {
	if j.err != nil {
		j.assertion.logFailure("JSONPath error at %q: %v", j.path, j.err)
		j.assertion.t.Errorf("JSONPath error at %q: %v", j.path, j.err)
		return j.assertion
	}
	if !j.exists {
		j.assertion.logFailure("expected JSON path %q to exist, but it was not found", j.path)
		j.assertion.t.Errorf("expected JSON path %q to exist, but it was not found", j.path)
	}
	return j.assertion
}

// EqString asserts that the JSON path matches the expected string.
func (j *JSONAssertionBuilder) EqString(want string) *AssertionBuilder {
	if j.err != nil {
		j.assertion.logFailure("JSONPath error at %q: %v", j.path, j.err)
		j.assertion.t.Errorf("JSONPath error at %q: %v", j.path, j.err)
		return j.assertion
	}
	if !j.exists {
		j.assertion.logFailure("expected JSON path %q to be %q, but path does not exist", j.path, want)
		j.assertion.t.Errorf("expected JSON path %q to be %q, but path does not exist", j.path, want)
		return j.assertion
	}
	got, ok := j.value.(string)
	if !ok {
		j.assertion.logFailure("expected JSON path %q to be string, got %T (%v)", j.path, j.value, j.value)
		j.assertion.t.Errorf("expected JSON path %q to be string, got %T (%v)", j.path, j.value, j.value)
		return j.assertion
	}
	if got != want {
		j.assertion.logFailure("expected JSON path %q to be %q, got %q", j.path, want, got)
		j.assertion.t.Errorf("expected JSON path %q to be %q, got %q", j.path, want, got)
	}
	return j.assertion
}

// EqBool asserts that the JSON path matches the expected boolean value.
func (j *JSONAssertionBuilder) EqBool(want bool) *AssertionBuilder {
	if j.err != nil {
		j.assertion.logFailure("JSONPath error at %q: %v", j.path, j.err)
		j.assertion.t.Errorf("JSONPath error at %q: %v", j.path, j.err)
		return j.assertion
	}
	if !j.exists {
		j.assertion.logFailure("expected JSON path %q to be %t, but path does not exist", j.path, want)
		j.assertion.t.Errorf("expected JSON path %q to be %t, but path does not exist", j.path, want)
		return j.assertion
	}
	got, ok := j.value.(bool)
	if !ok {
		j.assertion.logFailure("expected JSON path %q to be bool, got %T (%v)", j.path, j.value, j.value)
		j.assertion.t.Errorf("expected JSON path %q to be bool, got %T (%v)", j.path, j.value, j.value)
		return j.assertion
	}
	if got != want {
		j.assertion.logFailure("expected JSON path %q to be %t, got %t", j.path, want, got)
		j.assertion.t.Errorf("expected JSON path %q to be %t, got %t", j.path, want, got)
	}
	return j.assertion
}

// EqInt asserts that the JSON path matches the expected integer value.
func (j *JSONAssertionBuilder) EqInt(want int) *AssertionBuilder {
	if j.err != nil {
		j.assertion.logFailure("JSONPath error at %q: %v", j.path, j.err)
		j.assertion.t.Errorf("JSONPath error at %q: %v", j.path, j.err)
		return j.assertion
	}
	if !j.exists {
		j.assertion.logFailure("expected JSON path %q to be %d, but path does not exist", j.path, want)
		j.assertion.t.Errorf("expected JSON path %q to be %d, but path does not exist", j.path, want)
		return j.assertion
	}
	var got int
	switch v := j.value.(type) {
	case float64:
		got = int(v)
	case int:
		got = v
	default:
		j.assertion.logFailure("expected JSON path %q to be numeric, got %T (%v)", j.path, j.value, j.value)
		j.assertion.t.Errorf("expected JSON path %q to be numeric, got %T (%v)", j.path, j.value, j.value)
		return j.assertion
	}
	if got != want {
		j.assertion.logFailure("expected JSON path %q to be %d, got %d", j.path, want, got)
		j.assertion.t.Errorf("expected JSON path %q to be %d, got %d", j.path, want, got)
	}
	return j.assertion
}

// IsNull asserts that the JSON path value is null.
func (j *JSONAssertionBuilder) IsNull() *AssertionBuilder {
	if j.err != nil {
		j.assertion.logFailure("JSONPath error at %q: %v", j.path, j.err)
		j.assertion.t.Errorf("JSONPath error at %q: %v", j.path, j.err)
		return j.assertion
	}
	if !j.exists {
		j.assertion.logFailure("expected JSON path %q to be null, but path does not exist", j.path)
		j.assertion.t.Errorf("expected JSON path %q to be null, but path does not exist", j.path)
		return j.assertion
	}
	if j.value != nil {
		j.assertion.logFailure("expected JSON path %q to be null, got %T (%v)", j.path, j.value, j.value)
		j.assertion.t.Errorf("expected JSON path %q to be null, got %T (%v)", j.path, j.value, j.value)
	}
	return j.assertion
}

// IsArray asserts that the JSON path value is an array.
func (j *JSONAssertionBuilder) IsArray() *AssertionBuilder {
	if j.err != nil {
		j.assertion.logFailure("JSONPath error at %q: %v", j.path, j.err)
		j.assertion.t.Errorf("JSONPath error at %q: %v", j.path, j.err)
		return j.assertion
	}
	if !j.exists {
		j.assertion.logFailure("expected JSON path %q to be array, but path does not exist", j.path)
		j.assertion.t.Errorf("expected JSON path %q to be array, but path does not exist", j.path)
		return j.assertion
	}
	_, ok := j.value.([]any)
	if !ok {
		j.assertion.logFailure("expected JSON path %q to be array, got %T (%v)", j.path, j.value, j.value)
		j.assertion.t.Errorf("expected JSON path %q to be array, got %T (%v)", j.path, j.value, j.value)
	}
	return j.assertion
}

// HasLen asserts that the JSON path value has the expected length.
func (j *JSONAssertionBuilder) HasLen(want int) *AssertionBuilder {
	if j.err != nil {
		j.assertion.logFailure("JSONPath error at %q: %v", j.path, j.err)
		j.assertion.t.Errorf("JSONPath error at %q: %v", j.path, j.err)
		return j.assertion
	}
	if !j.exists {
		j.assertion.logFailure("expected JSON path %q to have len %d, but path does not exist", j.path, want)
		j.assertion.t.Errorf("expected JSON path %q to have len %d, but path does not exist", j.path, want)
		return j.assertion
	}
	var got int
	switch v := j.value.(type) {
	case []any:
		got = len(v)
	case map[string]any:
		got = len(v)
	case string:
		got = len(v)
	default:
		j.assertion.logFailure("expected JSON path %q to support len, got %T (%v)", j.path, j.value, j.value)
		j.assertion.t.Errorf("expected JSON path %q to support len, got %T (%v)", j.path, j.value, j.value)
		return j.assertion
	}
	if got != want {
		j.assertion.logFailure("expected JSON path %q to have len %d, got %d", j.path, want, got)
		j.assertion.t.Errorf("expected JSON path %q to have len %d, got %d", j.path, want, got)
	}
	return j.assertion
}

// pathComponent represents a segment of a dot-notation path.
type pathComponent struct {
	raw     string
	key     string
	isIndex bool
	index   int
}

// parsePath splits a path string into components, parsing array indices like [0].
func parsePath(path string) []pathComponent {
	var components []pathComponent
	parts := strings.Split(path, ".")
	for _, part := range parts {
		if part == "" {
			continue
		}

		currentKey := ""
		inBracket := false
		bracketVal := ""

		for i := 0; i < len(part); i++ {
			char := part[i]
			if char == '[' {
				if currentKey != "" && !inBracket {
					components = append(components, pathComponent{
						raw: currentKey,
						key: currentKey,
					})
					currentKey = ""
				}
				inBracket = true
				bracketVal = ""
			} else if char == ']' {
				if inBracket {
					idx, err := strconv.Atoi(bracketVal)
					if err == nil {
						components = append(components, pathComponent{
							raw:     "[" + bracketVal + "]",
							isIndex: true,
							index:   idx,
						})
					}
					inBracket = false
				}
			} else {
				if inBracket {
					bracketVal += string(char)
				} else {
					currentKey += string(char)
				}
			}
		}
		if currentKey != "" {
			components = append(components, pathComponent{
				raw: currentKey,
				key: currentKey,
			})
		}
	}
	return components
}

// evaluatePath evaluates a path of components on a parsed JSON object.
func evaluatePath(obj any, path string) (any, bool, error) {
	if path == "" {
		return nil, false, fmt.Errorf("empty path")
	}
	if !strings.HasPrefix(path, "$.") && path != "$" {
		return nil, false, fmt.Errorf("path must start with $.")
	}
	if path == "$" {
		return obj, true, nil
	}

	parts := parsePath(path[2:])
	current := obj

	for _, part := range parts {
		if part.isIndex {
			slice, ok := current.([]any)
			if !ok {
				return nil, false, fmt.Errorf("expected slice at path component %s, got %T", part.raw, current)
			}
			if part.index < 0 || part.index >= len(slice) {
				return nil, false, fmt.Errorf("index out of range [%d] (len=%d) at path component %s", part.index, len(slice), part.raw)
			}
			current = slice[part.index]
		} else {
			m, ok := current.(map[string]any)
			if !ok {
				return nil, false, fmt.Errorf("expected map at path component %q, got %T", part.key, current)
			}
			val, exists := m[part.key]
			if !exists {
				return nil, false, nil
			}
			current = val
		}
	}

	return current, true, nil
}

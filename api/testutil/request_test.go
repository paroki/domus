package testutil_test

import (
	"bytes"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/testutil"
)

func TestRequestBuilder_AllAssertions(t *testing.T) {
	app := fiber.New()

	app.Post("/echo", func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		customHeader := c.Get("X-Custom")
		contentType := c.Get("Content-Type")

		// Echo back request info and a complex nested JSON structure
		c.Set("X-Echo-Auth", authHeader)
		c.Set("X-Echo-Custom", customHeader)
		c.Set("X-Echo-Type", contentType)

		return c.Status(fiber.StatusCreated).JSON(map[string]any{
			"success": true,
			"data": []any{
				map[string]any{
					"id":      "usr_1",
					"email":   "alice@example.com",
					"active":  true,
					"age":     30,
					"profile": map[string]any{"role": "admin"},
					"tags":    []any{"staff", "primary"},
				},
				map[string]any{
					"id":      "usr_2",
					"email":   "bob@example.com",
					"active":  false,
					"age":     25,
					"profile": nil,
					"tags":    []any{},
				},
			},
			"meta": map[string]any{
				"pagination": map[string]any{
					"page":     1,
					"per_page": 20,
				},
			},
		})
	})

	body := map[string]any{"foo": "bar"}

	testutil.New(app).
		POST("/echo").
		WithJSON(body).
		WithHeader("X-Custom", "my-value").
		WithBearerToken("my-token").
		Expect(t).
		StatusCreated().
		Header("X-Echo-Auth").Eq("Bearer my-token").
		Header("X-Echo-Custom").Eq("my-value").
		Header("X-Echo-Type").Eq("application/json").
		Header("X-Echo-Custom").Exists().
		Header("X-Nonexistent").DoesNotExist().
		JSONPath("$.success").EqBool(true).
		JSONPath("$.data").IsArray().
		JSONPath("$.data").HasLen(2).
		JSONPath("$.data[0].id").EqString("usr_1").
		JSONPath("$.data[0].email").EqString("alice@example.com").
		JSONPath("$.data[0].active").EqBool(true).
		JSONPath("$.data[0].age").EqInt(30).
		JSONPath("$.data[0].profile.role").EqString("admin").
		JSONPath("$.data[0].tags").IsArray().
		JSONPath("$.data[0].tags").HasLen(2).
		JSONPath("$.data[0].tags[0]").EqString("staff").
		JSONPath("$.data[0].tags[1]").EqString("primary").
		JSONPath("$.data[1].id").EqString("usr_2").
		JSONPath("$.data[1].active").EqBool(false).
		JSONPath("$.data[1].profile").IsNull().
		JSONPath("$.data[1].tags").HasLen(0).
		JSONPath("$.meta.pagination.page").EqInt(1).
		JSONPath("$.meta.pagination.per_page").EqInt(20)
}

func TestRequestBuilder_WithBody(t *testing.T) {
	app := fiber.New()

	app.Put("/upload", func(c fiber.Ctx) error {
		body := c.Body()
		c.Set("Content-Type", "text/plain")
		return c.SendString("received: " + string(body))
	})

	buf := bytes.NewBufferString("hello world")

	testutil.New(app).
		PUT("/upload").
		WithBody(buf, "text/plain").
		Expect(t).
		StatusOK().
		Header("Content-Type").Eq("text/plain")
}

package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestRequestIDMiddlewareSetsHeader(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())
	app.Get("/test", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) })

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app test: %v", err)
	}
	defer resp.Body.Close()

	if resp.Header.Get(RequestIDHeader) == "" {
		t.Fatal("expected X-Request-ID header to be set")
	}
}

func TestRequestIDMiddlewarePreservesExisting(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString(c.Get(RequestIDHeader))
	})

	existingID := "custom-request-id-123"
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(RequestIDHeader, existingID)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app test: %v", err)
	}
	defer resp.Body.Close()

	if resp.Header.Get(RequestIDHeader) != existingID {
		t.Fatalf("expected preserved ID %s, got %s", existingID, resp.Header.Get(RequestIDHeader))
	}
}

func TestRequestIDStoredInLocals(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())

	var localID string
	app.Get("/test", func(c *fiber.Ctx) error {
		localID = c.Locals("request_id").(string)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	_, err := app.Test(req)
	if err != nil {
		t.Fatalf("app test: %v", err)
	}

	if localID == "" {
		t.Fatal("expected request_id to be stored in locals")
	}
}

func TestNewRIDGeneratesUnique(t *testing.T) {
	// Generate multiple IDs and ensure they're unique
	ids := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := newRID()
		if ids[id] {
			t.Fatalf("generated duplicate ID: %s", id)
		}
		ids[id] = true

		// Check length (base64url of 16 bytes = 22 chars)
		if len(id) != 22 {
			t.Fatalf("expected ID length 22, got %d", len(id))
		}
	}
}

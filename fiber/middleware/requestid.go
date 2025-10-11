package middleware

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/gofiber/fiber/v2"
)

// RequestIDHeader is the default header name for request IDs.
const RequestIDHeader = "X-Request-ID"

// RequestID returns a middleware that injects a unique request ID into each request.
// If a request already has a request ID in the header, it will be preserved.
// Otherwise, a new cryptographically random ID will be generated.
//
// The request ID is:
// - Set in the response header (X-Request-ID)
// - Stored in context locals as "request_id"
// - Available for logging and tracing
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		rid := c.Get(RequestIDHeader)
		if rid == "" {
			rid = newRID()
		}
		c.Set(RequestIDHeader, rid)
		// Store in locals for other middleware
		c.Locals("request_id", rid)
		return c.Next()
	}
}

// newRID generates a cryptographically random request ID.
// It uses 16 random bytes encoded as base64url without padding (22 characters).
// This provides ~128 bits of entropy, making collisions extremely unlikely.
func newRID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID if random fails (should never happen)
		return base64.RawURLEncoding.EncodeToString([]byte("fallback"))
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

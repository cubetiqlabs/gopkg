package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// AdminMiddleware returns a Fiber handler that validates the X-Admin-Secret header.
// This is useful for protecting admin endpoints that should only be accessible
// with a secret token.
//
// Example usage:
//
//	adminRoutes := app.Group("/admin", middleware.AdminMiddleware("my-secret-token"))
//	adminRoutes.Get("/users", listUsers)
//
// Security notes:
// - The secret should be strong and stored securely (environment variable, secrets manager)
// - Consider using this in combination with rate limiting
// - For production, consider more robust authentication (JWT, OAuth)
func AdminMiddleware(expected string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Reject if no secret is configured (safety check)
		if expected == "" {
			return fiber.ErrForbidden
		}
		
		secret := c.Get("X-Admin-Secret")
		if secret == "" || secret != expected {
			return fiber.ErrUnauthorized
		}
		
		return c.Next()
	}
}

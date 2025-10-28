package middleware

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// SecurityHeadersConfig defines configuration for security headers.
type SecurityHeadersConfig struct {
	// HSTSMaxAge sets the max-age for Strict-Transport-Security header (default: 31536000 = 1 year)
	HSTSMaxAge int

	// ContentSecurityPolicy defines the CSP header value
	// Default: "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'"
	ContentSecurityPolicy string

	// EnableXSSProtection enables X-XSS-Protection header (default: true)
	EnableXSSProtection bool

	// EnableFrameOptions enables X-Frame-Options header (default: true)
	EnableFrameOptions bool

	// EnableContentTypeNosniff enables X-Content-Type-Options header (default: true)
	EnableContentTypeNosniff bool
}

// SecurityHeaders returns a middleware that sets secure HTTP headers with default configuration.
// This helps protect against common web vulnerabilities including:
// - Clickjacking (X-Frame-Options)
// - MIME type sniffing (X-Content-Type-Options)
// - XSS attacks (Content-Security-Policy, X-XSS-Protection)
// - Man-in-the-middle attacks (Strict-Transport-Security)
//
// Example usage:
//
//	app.Use(middleware.SecurityHeaders())
func SecurityHeaders() fiber.Handler {
	return SecurityHeadersWithConfig(SecurityHeadersConfig{})
}

// SecurityHeadersWithConfig returns a security headers middleware with custom configuration.
//
// Example usage:
//
//	app.Use(middleware.SecurityHeadersWithConfig(middleware.SecurityHeadersConfig{
//	    HSTSMaxAge: 63072000, // 2 years
//	    ContentSecurityPolicy: "default-src 'self'",
//	}))
func SecurityHeadersWithConfig(cfg SecurityHeadersConfig) fiber.Handler {
	// Set defaults
	if cfg.HSTSMaxAge == 0 {
		cfg.HSTSMaxAge = 31536000 // 1 year in seconds
	}
	if cfg.ContentSecurityPolicy == "" {
		cfg.ContentSecurityPolicy = "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'"
	}

	// Default to enabled
	if !cfg.EnableXSSProtection {
		cfg.EnableXSSProtection = true
	}
	if !cfg.EnableFrameOptions {
		cfg.EnableFrameOptions = true
	}
	if !cfg.EnableContentTypeNosniff {
		cfg.EnableContentTypeNosniff = true
	}

	return func(c *fiber.Ctx) error {
		// HSTS: Force HTTPS connections
		// Only set over HTTPS to avoid browser warnings
		if c.Protocol() == "https" {
			c.Set("Strict-Transport-Security", "max-age="+strconv.Itoa(cfg.HSTSMaxAge)+"; includeSubDomains")
		}

		// CSP: Control resources the browser can load
		c.Set("Content-Security-Policy", cfg.ContentSecurityPolicy)

		// X-Frame-Options: Prevent clickjacking
		if cfg.EnableFrameOptions {
			c.Set("X-Frame-Options", "DENY")
		}

		// X-Content-Type-Options: Prevent MIME type sniffing
		if cfg.EnableContentTypeNosniff {
			c.Set("X-Content-Type-Options", "nosniff")
		}

		// X-XSS-Protection: Enable browser XSS filtering
		if cfg.EnableXSSProtection {
			c.Set("X-XSS-Protection", "1; mode=block")
		}

		// Referrer-Policy: Control referrer information
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy: Control browser features
		c.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		return c.Next()
	}
}

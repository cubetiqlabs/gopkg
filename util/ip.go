package util

import (
	"github.com/gofiber/fiber/v2"
)

// GetClientIP extracts the real client IP from various headers and fallbacks.
// Priority: CF-Connecting-IP > X-Real-IP > X-Forwarded-For > RemoteAddr
func GetClientIP(c *fiber.Ctx) string {
	// Cloudflare proxy: CF-Connecting-IP header contains the actual client IP
	cfConnectingIP := c.Get("CF-Connecting-IP")
	if cfConnectingIP != "" {
		return cfConnectingIP
	}

	// Standard reverse proxy header
	xRealIP := c.Get("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	// X-Forwarded-For can contain multiple IPs (client, proxy1, proxy2...)
	// The first IP is the original client
	clientIPs := c.IPs()
	if len(clientIPs) > 0 && clientIPs[0] != "" {
		return clientIPs[0]
	}

	// Fallback to Fiber's IP() method which uses RemoteAddr
	return c.IP()
}

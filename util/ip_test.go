package util

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name        string
		headers     map[string]string
		expectedIP  string
		description string
	}{
		{
			name: "CF-Connecting-IP header",
			headers: map[string]string{
				"CF-Connecting-IP": "203.0.113.1",
				"X-Real-IP":        "198.51.100.1",
				"X-Forwarded-For":  "192.0.2.1",
			},
			expectedIP:  "203.0.113.1",
			description: "CloudFlare IP should have highest priority",
		},
		{
			name: "X-Real-IP header",
			headers: map[string]string{
				"X-Real-IP":       "198.51.100.1",
				"X-Forwarded-For": "192.0.2.1",
			},
			expectedIP:  "198.51.100.1",
			description: "X-Real-IP should be second priority",
		},
		{
			name: "X-Forwarded-For single IP",
			headers: map[string]string{
				"X-Forwarded-For": "192.0.2.1",
			},
			expectedIP:  "192.0.2.1",
			description: "X-Forwarded-For with single IP",
		},
		{
			name: "X-Forwarded-For multiple IPs",
			headers: map[string]string{
				"X-Forwarded-For": "192.0.2.1, 198.51.100.1, 203.0.113.1",
			},
			expectedIP:  "192.0.2.1",
			description: "X-Forwarded-For should return first IP",
		},
		{
			name: "X-Forwarded-For with spaces",
			headers: map[string]string{
				"X-Forwarded-For": "  192.0.2.1  ,  198.51.100.1  ",
			},
			expectedIP:  "192.0.2.1",
			description: "X-Forwarded-For should trim spaces",
		},
		{
			name:        "Direct connection",
			headers:     map[string]string{},
			expectedIP:  "0.0.0.0",
			description: "Should use context IP when no proxy headers",
		},
		{
			name: "IPv6 address",
			headers: map[string]string{
				"X-Real-IP": "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			},
			expectedIP:  "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			description: "Should handle IPv6 addresses",
		},
		{
			name: "Empty X-Forwarded-For",
			headers: map[string]string{
				"X-Forwarded-For": "",
			},
			expectedIP:  "0.0.0.0",
			description: "Empty X-Forwarded-For should fallback to context IP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			
			var resultIP string
			app.Get("/test", func(c *fiber.Ctx) error {
				resultIP = GetClientIP(c)
				return c.SendString("OK")
			})

			req := httptest.NewRequest("GET", "/test", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			
			_, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedIP, resultIP, tt.description)
		})
	}
}

func TestGetClientIP_HeaderPriority(t *testing.T) {
	app := fiber.New()
	
	var resultIP string
	app.Get("/test", func(c *fiber.Ctx) error {
		resultIP = GetClientIP(c)
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	// Set all headers to verify priority
	req.Header.Set("CF-Connecting-IP", "1.1.1.1")
	req.Header.Set("X-Real-IP", "2.2.2.2")
	req.Header.Set("X-Forwarded-For", "3.3.3.3, 4.4.4.4")
	
	_, err := app.Test(req)
	assert.NoError(t, err)
	
	// CloudFlare header should win
	assert.Equal(t, "1.1.1.1", resultIP, "Expected CF-Connecting-IP to have highest priority")
}


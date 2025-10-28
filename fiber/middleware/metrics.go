package middleware

import (
	"strconv"
	"time"

	"github.com/cubetiqlabs/gopkg/contextx"
	"github.com/cubetiqlabs/gopkg/metrics"
	"github.com/gofiber/fiber/v2"
)

// Metrics returns a Fiber middleware that collects request metrics.
// It tracks:
// - Total requests
// - Request duration (avg, sum, count)
// - Labeled metrics by method, path, status, and optionally tenant
//
// Example usage:
//
//	reg := metrics.NewRegistry()
//	app.Use(middleware.Metrics(reg))
//
//	// Expose metrics endpoint
//	app.Get("/metrics", func(c *fiber.Ctx) error {
//	    c.Set("Content-Type", "text/plain")
//	    return c.SendString(reg.RenderPrometheus())
//	})
func Metrics(reg *metrics.Registry) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Record metrics
		durMs := time.Since(start).Milliseconds()
		reg.RequestsTotal.Inc()
		reg.RequestDuration.Observe(durMs)

		// Extract tenant if available
		tenantID, _ := contextx.TenantID(c.UserContext())

		// Record labeled metric
		reg.IncLabeled("http_requests", map[string]string{
			"method": c.Method(),
			"path":   c.Route().Path,
			"status": strconv.Itoa(c.Response().StatusCode()),
			"tenant": tenantID,
		})

		return err
	}
}

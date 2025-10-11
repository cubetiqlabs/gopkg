# Fiber Middleware

Production-ready middleware components for [Fiber](https://gofiber.io/) web framework.

## Overview

This package provides a collection of battle-tested middleware components extracted from real-world production applications. Each middleware is designed to be composable, configurable, and performant.

## Available Middleware

- **[RequestID](#requestid)** - Generate unique request IDs for tracing
- **[Admin](#admin)** - Protect admin routes with secret-based authentication
- **[Security](#security)** - Add security headers (HSTS, CSP, XSS protection, etc.)
- **[Error](#error)** - Centralized error handling with security-conscious responses
- **[AccessLog](#accesslog)** - Structured access logging with request/response details
- **[Metrics](#metrics)** - Collect HTTP metrics (requests, duration, status codes)
- **[RateLimit](#ratelimit)** - Token bucket rate limiter with automatic cleanup

## Installation

```bash
go get github.com/cubetiqlabs/gopkg/fiber/middleware
```

## Middleware Details

### RequestID

Generates a unique, cryptographically random request ID for each request. IDs are 22 characters long (base64url-encoded 16 random bytes) providing ~128 bits of entropy.

**Features:**
- Preserves existing request IDs from upstream proxies
- Adds `X-Request-ID` header to responses
- Stores ID in context locals as `request_id`
- Suitable for distributed tracing

**Usage:**

```go
import (
    "github.com/cubetiqlabs/gopkg/fiber/middleware"
    "github.com/gofiber/fiber/v2"
)

app := fiber.New()
app.Use(middleware.RequestID())

app.Get("/api/users", func(c *fiber.Ctx) error {
    // Access request ID from locals
    requestID := c.Locals("request_id").(string)
    
    // Use it in logs, error responses, etc.
    return c.JSON(fiber.Map{
        "request_id": requestID,
        "users": []string{"alice", "bob"},
    })
})
```

**Response Headers:**
```
X-Request-ID: 1a2b3c4d5e6f7g8h9i0j1k
```

---

### Admin

Protects routes with secret-based authentication. Useful for admin panels, internal tools, or privileged operations.

**Features:**
- Simple secret-based authentication
- Configurable header name (default: `X-Admin-Secret`)
- Returns 401 Unauthorized on missing/invalid secrets
- No performance overhead for public routes

**Usage:**

```go
app := fiber.New()

// Apply to specific route groups
admin := app.Group("/admin")
admin.Use(middleware.Admin("your-secret-here"))

admin.Get("/dashboard", func(c *fiber.Ctx) error {
    return c.SendString("Admin Dashboard")
})

admin.Post("/users", func(c *fiber.Ctx) error {
    // Only accessible with correct X-Admin-Secret header
    return c.SendString("User created")
})
```

**Custom Configuration:**

```go
app.Use(middleware.AdminWithConfig(middleware.AdminConfig{
    Secret: "super-secret-key",
    Header: "X-Internal-Token", // Custom header name
}))
```

**Request Headers:**
```bash
curl -H "X-Admin-Secret: your-secret-here" \
     https://api.example.com/admin/dashboard
```

---

### Security

Adds comprehensive security headers to protect against common web vulnerabilities.

**Features:**
- HSTS (HTTP Strict Transport Security)
- Content Security Policy (CSP)
- XSS Protection
- Content Type Options (prevent MIME sniffing)
- Frame Options (clickjacking protection)
- Referrer Policy
- Permissions Policy
- Customizable for each header

**Usage:**

```go
app := fiber.New()

// Use with defaults
app.Use(middleware.SecurityHeaders())

app.Get("/", func(c *fiber.Ctx) error {
    return c.SendString("Hello, World!")
})
```

**Custom Configuration:**

```go
app.Use(middleware.SecurityHeadersWithConfig(middleware.SecurityHeadersConfig{
    HSTS: "max-age=63072000; includeSubDomains; preload",
    CSP: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'",
    XSSProtection: "1; mode=block",
    ContentTypeNosniff: "nosniff",
    FrameOptions: "SAMEORIGIN",
    ReferrerPolicy: "strict-origin-when-cross-origin",
    PermissionsPolicy: "geolocation=(), microphone=(), camera=()",
}))
```

**Response Headers:**
```
Strict-Transport-Security: max-age=31536000; includeSubDomains
Content-Security-Policy: default-src 'self'
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Referrer-Policy: no-referrer
Permissions-Policy: geolocation=(), microphone=(), camera=()
```

---

### Error

Centralized error handling middleware that intercepts errors and returns consistent, security-conscious responses.

**Features:**
- Consistent error response format
- Security-conscious (hides internal error details in production)
- Fiber error support (status codes, messages)
- Structured logging with request ID correlation
- Configurable error masking

**Usage:**

```go
import (
    "github.com/cubetiqlabs/gopkg/logging"
    "github.com/cubetiqlabs/gopkg/fiber/middleware"
)

// Initialize logger
logger := logging.Init("production")

app := fiber.New()
app.Use(middleware.RequestID())
app.Use(middleware.ErrorHandlerWithLogger(logger))

app.Get("/users/:id", func(c *fiber.Ctx) error {
    // Return Fiber errors with specific status codes
    return fiber.NewError(fiber.StatusNotFound, "user not found")
})
```

**Error Response Format:**

```json
{
    "error": "user not found",
    "request_id": "1a2b3c4d5e6f7g8h9i0j1k"
}
```

**Production Mode:**
- 500 errors show: `"error": "internal server error"`
- Request ID always included for debugging

**Development Mode:**
- Full error details exposed for easier debugging

---

### AccessLog

Structured access logging middleware that records detailed information about each request/response.

**Features:**
- Structured JSON logs with Zap logger
- Records: method, path, status, duration, IP, user agent, request ID
- Configurable log level (info for 2xx/3xx, warn for 4xx, error for 5xx)
- Integration with request ID middleware
- Sub-millisecond precision timing

**Usage:**

```go
import (
    "github.com/cubetiqlabs/gopkg/logging"
    "github.com/cubetiqlabs/gopkg/fiber/middleware"
)

logger := logging.Init("production")

app := fiber.New()
app.Use(middleware.RequestID())
app.Use(middleware.AccessLog(logger))

app.Get("/api/users", func(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{"users": []string{"alice", "bob"}})
})
```

**Log Output:**

```json
{
    "level": "info",
    "ts": 1696956000.123456,
    "msg": "access",
    "method": "GET",
    "path": "/api/users",
    "status": 200,
    "duration_ms": 45.67,
    "ip": "203.0.113.1",
    "user_agent": "Mozilla/5.0...",
    "request_id": "1a2b3c4d5e6f7g8h9i0j1k"
}
```

---

### Metrics

Collects HTTP request metrics for monitoring and observability.

**Features:**
- Total request count
- Request duration histogram
- Per-endpoint metrics with labels (method, path, status)
- Per-tenant metrics (if tenant context available)
- Prometheus-compatible output
- Thread-safe atomic operations

**Usage:**

```go
import (
    "github.com/cubetiqlabs/gopkg/metrics"
    "github.com/cubetiqlabs/gopkg/fiber/middleware"
)

registry := metrics.NewRegistry()

app := fiber.New()
app.Use(middleware.Metrics(registry))

// Expose metrics endpoint
app.Get("/metrics", func(c *fiber.Ctx) error {
    c.Set("Content-Type", "text/plain")
    return c.SendString(registry.RenderPrometheus())
})
```

**Collected Metrics:**

- `http_requests_total` - Total HTTP requests
- `http_request_duration_ms_avg` - Average request duration
- `http_requests{method="GET",path="/api/users",status="200"}` - Labeled per-endpoint metrics
- `http_requests{tenant="<id>"}` - Per-tenant metrics (when tenant context exists)

**Prometheus Output:**

```
http_requests_total 12345
http_request_duration_ms_avg 45.67
http_requests{method="GET",path="/api/users",status="200"} 5678
uptime_seconds 3600
```

---

### RateLimit

Token bucket rate limiter with configurable limits and automatic cleanup.

**Features:**
- Per-key rate limiting (tenant, API key, IP, etc.)
- Token bucket algorithm with burst capacity
- Dynamic burst (automatically set to half of rate)
- Automatic bucket cleanup to prevent memory exhaustion
- Retry-After header for rejected requests
- Metrics integration (rate_allowed_total, rate_rejected_total)

**Usage:**

```go
import (
    "github.com/cubetiqlabs/gopkg/metrics"
    "github.com/cubetiqlabs/gopkg/fiber/middleware"
)

registry := metrics.NewRegistry()

app := fiber.New()

// Global rate limit: 100 requests per minute per IP
app.Use(middleware.RateLimit(
    middleware.RateLimitConfig{
        RequestsPerMinute: 100,
        KeyGenerator: func(c *fiber.Ctx) string {
            return c.IP() // Rate limit by IP address
        },
        Registry: registry,
    },
))
```

**Per-Tenant Rate Limiting:**

```go
app.Use(middleware.RateLimit(
    middleware.RateLimitConfig{
        RequestsPerMinute: 1000,
        KeyGenerator: func(c *fiber.Ctx) string {
            // Assumes tenant ID is stored in context
            tenantID := c.Locals("tenant_id").(string)
            return tenantID
        },
        Registry: registry,
    },
))
```

**Response Headers (when rate limited):**

```
HTTP/1.1 429 Too Many Requests
Retry-After: 45
```

**Burst Capacity:**
- Burst is automatically set to `rate / 2`
- Example: 100 req/min → burst of 50 requests
- Allows handling traffic spikes while maintaining rate limit

**Memory Management:**
- Inactive buckets are cleaned up every 5 minutes
- Bucket considered stale after 15 minutes of inactivity
- Maximum 10,000 buckets to prevent memory exhaustion

---

## Complete Example

Here's a complete example combining multiple middleware:

```go
package main

import (
    "github.com/cubetiqlabs/gopkg/fiber/middleware"
    "github.com/cubetiqlabs/gopkg/logging"
    "github.com/cubetiqlabs/gopkg/metrics"
    "github.com/gofiber/fiber/v2"
)

func main() {
    // Initialize dependencies
    logger := logging.Init("production")
    registry := metrics.NewRegistry()

    // Create Fiber app
    app := fiber.New(fiber.Config{
        ErrorHandler: middleware.ErrorHandlerWithLogger(logger),
    })

    // Apply global middleware (order matters!)
    app.Use(middleware.RequestID())
    app.Use(middleware.SecurityHeaders())
    app.Use(middleware.AccessLog(logger))
    app.Use(middleware.Metrics(registry))
    app.Use(middleware.RateLimit(middleware.RateLimitConfig{
        RequestsPerMinute: 100,
        KeyGenerator: func(c *fiber.Ctx) string {
            return c.IP()
        },
        Registry: registry,
    }))

    // Public routes
    app.Get("/", func(c *fiber.Ctx) error {
        return c.SendString("Hello, World!")
    })

    app.Get("/api/users", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "users": []string{"alice", "bob", "charlie"},
        })
    })

    // Metrics endpoint
    app.Get("/metrics", func(c *fiber.Ctx) error {
        c.Set("Content-Type", "text/plain")
        return c.SendString(registry.RenderPrometheus())
    })

    // Admin routes (protected)
    admin := app.Group("/admin")
    admin.Use(middleware.Admin("super-secret-key"))
    
    admin.Get("/dashboard", func(c *fiber.Ctx) error {
        return c.SendString("Admin Dashboard")
    })

    // Start server
    logger.Info("Starting server on :8080")
    if err := app.Listen(":8080"); err != nil {
        logger.Fatal("Failed to start server", "error", err)
    }
}
```

## Middleware Order

For optimal functionality, apply middleware in this order:

1. **RequestID** - Generate request IDs first for tracing
2. **Security** - Set security headers early
3. **AccessLog** - Log after request ID is available
4. **Metrics** - Collect metrics for all requests
5. **RateLimit** - Apply rate limiting before business logic
6. **Error** - Handle errors last (or use as ErrorHandler in Fiber config)

## Testing

All middleware components have comprehensive test coverage:

```bash
cd fiber/middleware
go test -v
```

## Performance

All middleware is designed for production use with minimal performance overhead:

- **RequestID**: < 1μs per request (crypto random generation)
- **Admin**: < 100ns per request (simple string comparison)
- **Security**: < 500ns per request (header writes)
- **AccessLog**: < 10μs per request (structured logging)
- **Metrics**: < 2μs per request (atomic counters)
- **RateLimit**: < 5μs per request (token bucket with atomic operations)

## License

MIT License - See LICENSE file for details.

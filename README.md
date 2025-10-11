# gopkg

Reusable Go packages for building web applications with Fiber and standard tooling.

## Overview

This repository contains a collection of production-ready, reusable Go packages extracted from real-world applications. All packages are designed to work seamlessly with [GoFiber](https://gofiber.io/) but many are framework-agnostic.

## Installation

```bash
go get github.com/cubetiqlabs/gopkg
```

## Packages

### Middleware (`fiber/middleware`)

Production-ready Fiber middleware for common use cases:

- **`requestid`** - Request ID generation and injection
- **`accesslog`** - Structured access logging with Zap
- **`error`** - Centralized error handling with security-conscious responses
- **`security`** - Security headers (HSTS, CSP, X-Frame-Options, etc.)
- **`ratelimit`** - Token bucket rate limiter with per-tenant overrides
- **`admin`** - Admin secret authentication
- **`metrics`** - Prometheus-style metrics collection

### Context Utilities (`contextx`)

Type-safe context value management:

- Tenant ID injection/extraction
- Application ID handling
- API key actor tracking
- Combined auth values

### Utilities (`util`)

Common utilities for Fiber applications:

- **`error.go`** - Fiber error helpers (NotFoundError, BadRequestError, etc.)
- **`ip.go`** - Client IP detection (CloudFlare, X-Real-IP, X-Forwarded-For)

### Logging (`logging`)

Structured logging with Zap:

- Global logger initialization
- Context-aware logging
- Configurable log levels

### Metrics (`metrics`)

Lightweight Prometheus-compatible metrics:

- Counters and histograms
- Labeled metrics
- Prometheus text format export

### Models (`model`)

Common data models:

- **`problemdetails`** - RFC 7807 Problem Details
- **`validation`** - Validation error structures

## Quick Start

### Request ID Middleware

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/cubetiqlabs/gopkg/fiber/middleware"
)

func main() {
    app := fiber.New()
    
    // Add request ID to all requests
    app.Use(middleware.RequestID())
    
    app.Get("/", func(c *fiber.Ctx) error {
        // Request ID is available in header
        rid := c.Get("X-Request-ID")
        return c.SendString("Request ID: " + rid)
    })
    
    app.Listen(":3000")
}
```

### Error Handling

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/cubetiqlabs/gopkg/fiber/middleware"
    "github.com/cubetiqlabs/gopkg/util"
)

func main() {
    app := fiber.New(fiber.Config{
        ErrorHandler: middleware.ErrorHandler(),
    })
    
    app.Get("/user/:id", func(c *fiber.Ctx) error {
        id := c.Params("id")
        
        // Use helper for common errors
        if id == "" {
            return util.BadRequestError("user ID required")
        }
        
        // Internal errors are logged but not exposed to clients
        return nil
    })
    
    app.Listen(":3000")
}
```

### Rate Limiting

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/cubetiqlabs/gopkg/fiber/middleware"
    "github.com/cubetiqlabs/gopkg/metrics"
)

func main() {
    app := fiber.New()
    
    reg := metrics.NewRegistry()
    limiter := middleware.NewRateLimiter(600) // 600 req/min
    
    app.Use(middleware.RateLimitMiddleware(limiter, reg))
    
    app.Get("/api/data", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{"status": "ok"})
    })
    
    app.Listen(":3000")
}
```

### Context Values

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/cubetiqlabs/gopkg/contextx"
)

func main() {
    app := fiber.New()
    
    // Middleware to inject tenant ID
    app.Use(func(c *fiber.Ctx) error {
        tenantID := c.Get("X-Tenant-ID")
        ctx := contextx.WithTenant(c.UserContext(), tenantID)
        c.SetUserContext(ctx)
        return c.Next()
    })
    
    app.Get("/data", func(c *fiber.Ctx) error {
        // Extract tenant ID
        tenantID, ok := contextx.TenantID(c.UserContext())
        if !ok {
            return fiber.ErrUnauthorized
        }
        
        return c.JSON(fiber.Map{"tenant": tenantID})
    })
    
    app.Listen(":3000")
}
```

### Structured Logging

```go
package main

import (
    "github.com/cubetiqlabs/gopkg/logging"
    "go.uber.org/zap"
)

func main() {
    // Initialize logger
    logger, err := logging.Init("info", false)
    if err != nil {
        panic(err)
    }
    defer logger.Sync()
    
    // Use logger
    logger.Info("application started",
        zap.String("port", "3000"),
        zap.String("env", "production"),
    )
}
```

### Metrics Collection

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/cubetiqlabs/gopkg/metrics"
    "github.com/cubetiqlabs/gopkg/fiber/middleware"
)

func main() {
    app := fiber.New()
    
    // Create metrics registry
    reg := metrics.NewRegistry()
    
    // Add metrics middleware
    app.Use(middleware.Metrics(reg))
    
    // Expose metrics endpoint
    app.Get("/metrics", func(c *fiber.Ctx) error {
        c.Set("Content-Type", "text/plain")
        return c.SendString(reg.RenderPrometheus())
    })
    
    app.Get("/api/data", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{"status": "ok"})
    })
    
    app.Listen(":3000")
}
```

## Design Principles

1. **Framework Agnostic Where Possible** - Core utilities work without Fiber
2. **Security First** - Error handling prevents information disclosure
3. **Production Ready** - Battle-tested in real applications
4. **Minimal Dependencies** - Only essential dependencies
5. **Well Tested** - Comprehensive test coverage
6. **Documented** - Clear documentation and examples

## Package Organization

```
gopkg/
├── fiber/              # Fiber-specific packages
│   └── middleware/     # Fiber middleware
├── contextx/           # Context utilities (framework-agnostic)
├── util/              # General utilities
├── logging/           # Logging utilities
├── metrics/           # Metrics collection
└── model/             # Common data models
```

## Testing

```bash
go test ./...
```

## Contributing

Contributions are welcome! Please ensure:

1. All tests pass
2. Code is well documented
3. Examples are provided
4. Follows Go best practices

## License

MIT License - See LICENSE file for details

## Related Projects

- [TinyDB](https://github.com/cubetiqlabs/tinydb) - Document database built with these packages

## Roadmap

- [ ] HTTP client utilities
- [ ] Database helpers
- [ ] Configuration management
- [ ] Pagination utilities
- [ ] Storage abstractions (S3, GCS, local)
- [ ] Cache abstractions (Redis, in-memory)
- [ ] Background job processing
- [ ] Email/notification helpers

## Support

For issues, questions, or contributions, please visit [GitHub Issues](https://github.com/cubetiqlabs/gopkg/issues).

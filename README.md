# gopkg

Reusable Go packages for building web applications with Fiber v3 and standard tooling.

[![Go Reference](https://pkg.go.dev/badge/github.com/cubetiqlabs/gopkg.svg)](https://pkg.go.dev/github.com/cubetiqlabs/gopkg)

## Overview

This repository contains a collection of production-ready, reusable Go packages extracted from real-world applications. All packages are designed to work seamlessly with [GoFiber v3](https://gofiber.io/) but many are framework-agnostic.

**Requirements:** Go 1.27+ and Fiber v3 (`github.com/gofiber/fiber/v3`).

> **Migrating from Fiber v2?** All Fiber-facing packages now import `github.com/gofiber/fiber/v3`, handlers use `func(c fiber.Ctx) error`, and `c.UserContext()` / `c.SetUserContext()` are replaced by `c.Context()` / `c.SetContext()`.

## Installation

```bash
go get github.com/cubetiqlabs/gopkg
```

## Packages

| Package | Description |
| --- | --- |
| [`config`](#configuration-config) | Viper-based, type-safe YAML configuration |
| [`fiber/middleware`](#middleware-fibermiddleware) | Request ID, access log, errors, security, rate limiting, metrics |
| [`contextx`](#context-utilities-contextx) | Type-safe tenant/app/API-key context values |
| [`util`](#utilities-util) | Fiber error helpers, client IP detection, parsing helpers |
| [`types`](#shared-types-types) | Shared data types such as `DateRange` |
| [`logging`](#structured-logging-logging) | Zap logger initialization and context-aware logging |
| [`metrics`](#metrics-metrics) | Prometheus-compatible counters and histograms |

### Configuration (`config`)

Production-ready YAML configuration management with Viper:

- **Type-safe getters** - String, int, float, bool, duration, slice, and map types
- **Multi-environment support** - Load environment-specific configs (`config.production.yaml`)
- **Environment variable overrides** - Auto-bind with configurable prefix
- **Global singleton** - Optional global config instance via `SetGlobal` / `Global`
- **Custom loaders** - Extensible for custom config sources
- **Defaults & required values** - `GetOrDefault`, `MustGet`, and friends
- **File watching** - Hot reload callbacks via `WatchConfig` / `Watch`
- **Thread-safe** - Built-in RWMutex for concurrent access

See [`config/README.md`](config/README.md) and [`docs/config/`](docs/config/) for the full guide.

### Middleware (`fiber/middleware`)

Production-ready Fiber v3 middleware for common use cases:

- **`requestid`** - Request ID generation and injection (`RequestID()`)
- **`accesslog`** - Structured access logging with Zap (`AccessLogWithConfig`)
- **`error`** - Centralized error handling with security-conscious responses (`ErrorHandler()`)
- **`security`** - Security headers, HSTS, CSP, X-Frame-Options, etc. (`SecurityHeaders()`)
- **`ratelimit`** - Token bucket rate limiter with per-key overrides (`NewRateLimiter`, `RateLimitMiddleware`)
- **`admin`** - Admin secret authentication (`AdminMiddleware`)
- **`metrics`** - Prometheus-style metrics collection (`Metrics`)

See [`fiber/middleware/README.md`](fiber/middleware/README.md) for details.

### Context Utilities (`contextx`)

Type-safe context value management:

- Tenant ID injection/extraction (`WithTenant`, `TenantID`)
- Application ID handling (`WithApplication`, `AppID`)
- API key actor tracking (`WithAPIKeyPrefix`, `APIKeyActor`)
- Combined tenant auth values (`WithTenantAuthValues`, `TenantAuth`)

### Utilities (`util`)

Common utilities for Fiber v3 applications:

- **`error.go`** - Fiber error helpers (`NotFoundError`, `BadRequestError`, `UnauthorizedError`, `InternalServerErr`, etc.)
- **`ip.go`** - Client IP detection (Cloudflare, `X-Real-IP`, `X-Forwarded-For`)
- **`parser.go`** - Configuration-friendly parsers:
  - `ParseDuration` - Parses durations with `ns`, `us`/`µs`, `ms`, `s`, `m`, `h`, `d`, `w` units
  - `ParseDateRange` - Parses `YYYY-MM-DD` ranges into a `types.DateRange`

### Shared Types (`types`)

Common data types shared across packages:

- **`DateRange`** - Start/end date pair with JSON tags (`start_date`, `end_date`)

### Logging (`logging`)

Structured logging with Zap:

- Global logger initialization (`Init`)
- Context-aware logging (`WithContext`, `FromContext`)
- Package-level helpers (`Info`, `Debug`, `Warn`, `Error`, `Fatal` and `*f` variants)
- Configurable log levels and development mode

### Metrics (`metrics`)

Lightweight Prometheus-compatible metrics:

- Counters and histograms
- Labeled metrics (`IncLabeled`, `AddLabeled`)
- Prometheus text format export
- Registry reset for tests

## Quick Start

### Configuration Management

```go
package main

import (
    "github.com/cubetiqlabs/gopkg/config"
)

func main() {
    // Initialize config
    cfg, err := config.New(&config.Options{
        ConfigPath: "./config",
        Env:        "production",
        EnvPrefix:  "APP",
    })
    if err != nil {
        panic(err)
    }

    // Type-safe access
    port := cfg.GetIntOrDefault("server.port", 3000)
    host := cfg.GetStringOrDefault("server.host", "localhost")
    _ = host

    // Or use structured config
    type ServerConfig struct {
        Host string `mapstructure:"host"`
        Port int    `mapstructure:"port"`
    }
    var server ServerConfig
    if err := cfg.UnmarshalKey("server", &server); err != nil {
        panic(err)
    }

    // Make globally available
    config.SetGlobal(cfg)

    // Access from anywhere
    _ = config.Global().GetInt("server.port")
    _ = port
}
```

### Request ID Middleware

```go
package main

import (
    "github.com/gofiber/fiber/v3"
    "github.com/cubetiqlabs/gopkg/fiber/middleware"
)

func main() {
    app := fiber.New()

    // Add request ID to all requests
    app.Use(middleware.RequestID())

    app.Get("/", func(c fiber.Ctx) error {
        // Request ID is available in the response header and locals
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
    "github.com/gofiber/fiber/v3"
    "github.com/cubetiqlabs/gopkg/fiber/middleware"
    "github.com/cubetiqlabs/gopkg/util"
)

func main() {
    app := fiber.New(fiber.Config{
        ErrorHandler: middleware.ErrorHandler(),
    })

    app.Get("/user/:id", func(c fiber.Ctx) error {
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
    "github.com/gofiber/fiber/v3"
    "github.com/cubetiqlabs/gopkg/fiber/middleware"
    "github.com/cubetiqlabs/gopkg/metrics"
)

func main() {
    app := fiber.New()

    reg := metrics.NewRegistry()
    limiter := middleware.NewRateLimiter(600) // 600 req/min

    app.Use(middleware.RateLimitMiddleware(limiter, reg))

    app.Get("/api/data", func(c fiber.Ctx) error {
        return c.JSON(fiber.Map{"status": "ok"})
    })

    app.Listen(":3000")
}
```

### Context Values

```go
package main

import (
    "github.com/gofiber/fiber/v3"
    "github.com/cubetiqlabs/gopkg/contextx"
)

func main() {
    app := fiber.New()

    // Middleware to inject tenant ID
    app.Use(func(c fiber.Ctx) error {
        tenantID := c.Get("X-Tenant-ID")
        ctx := contextx.WithTenant(c.Context(), tenantID)
        c.SetContext(ctx)
        return c.Next()
    })

    app.Get("/data", func(c fiber.Ctx) error {
        // Extract tenant ID
        tenantID, ok := contextx.TenantID(c.Context())
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
    "github.com/gofiber/fiber/v3"
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
    app.Get("/metrics", func(c fiber.Ctx) error {
        c.Set("Content-Type", "text/plain")
        return c.SendString(reg.RenderPrometheus())
    })

    app.Get("/api/data", func(c fiber.Ctx) error {
        return c.JSON(fiber.Map{"status": "ok"})
    })

    app.Listen(":3000")
}
```

### Parsing Helpers

```go
package main

import (
    "fmt"

    "github.com/cubetiqlabs/gopkg/util"
)

func main() {
    // Supports ns, us/µs, ms, s, m, h, d (day), w (week)
    ttl, err := util.ParseDuration("1.5h")
    if err != nil {
        panic(err)
    }
    fmt.Println(ttl) // 1h30m0s

    // Parse a date range (inclusive end-of-day when includeTime is true)
    r, err := util.ParseDateRange("2026-01-01", "2026-01-31", true)
    if err != nil {
        panic(err)
    }
    fmt.Println(r.StartDate, r.EndDate)
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
├── config/             # Viper-based configuration management
├── contextx/           # Context utilities (framework-agnostic)
├── docs/
│   └── config/         # In-depth configuration guides
├── fiber/
│   └── middleware/     # Fiber v3 middleware
├── logging/            # Zap logging utilities
├── metrics/            # Metrics collection
├── types/              # Shared data types
└── util/               # General utilities (errors, IP, parsers)
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

- [x] Configuration management
- [x] Fiber v3 support
- [x] Context utilities
- [x] Structured logging and metrics
- [x] Parsing helpers (durations, date ranges)
- [ ] HTTP client utilities
- [ ] Database helpers
- [ ] Pagination utilities
- [ ] Storage abstractions (S3, GCS, local)
- [ ] Cache abstractions (Redis, in-memory)
- [ ] Background job processing
- [ ] Email/notification helpers

## Support

For issues, questions, or contributions, please visit [GitHub Issues](https://github.com/cubetiqlabs/gopkg/issues).

# CHANGELOG

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of gopkg - reusable Go utilities for web applications
- `contextx` package: Type-safe context utilities for multi-tenant applications
- `util` package: Fiber error helpers and client IP detection
- `logging` package: Zap logger initialization and context integration
- `metrics` package: Prometheus-compatible metrics collection
- `fiber/middleware` package: Production-ready Fiber middleware
  - RequestID: Cryptographically random request ID generation
  - Admin: Secret-based authentication for admin routes
  - Security: Comprehensive security headers (HSTS, CSP, XSS, etc.)
  - Error: Centralized error handling with structured responses
  - AccessLog: Structured access logging with Zap
  - Metrics: HTTP metrics collection (requests, duration, labels)
  - RateLimit: Token bucket rate limiter with automatic cleanup

### Test Coverage
- `contextx`: 96.9% coverage
- `metrics`: 100.0% coverage
- `util`: 76.0% coverage
- `fiber/middleware`: 5.8% coverage (basic RequestID tests)

## [0.1.0] - 2025-10-11

### Added
- Initial project structure
- Core packages extracted from TinyDB production application
- Comprehensive documentation and usage examples
- MIT License

[Unreleased]: https://github.com/cubetiqlabs/gopkg/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/cubetiqlabs/gopkg/releases/tag/v0.1.0

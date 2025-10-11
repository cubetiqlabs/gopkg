package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// AccessLogConfig defines configuration for access logging.
type AccessLogConfig struct {
	// Logger is the zap logger instance (required)
	Logger *zap.Logger
	
	// LevelResolver determines log level based on status code and error
	// Default: 2xx/3xx = Info, 4xx = Warn, 5xx = Error
	LevelResolver func(status int, err error) zapcore.Level
	
	// IncludeHeaders list of headers to include in logs (case-insensitive)
	// Example: []string{"X-Request-ID", "User-Agent"}
	IncludeHeaders []string
	
	// Skip is a function to skip logging for certain requests
	// Example: func(c *fiber.Ctx) bool { return c.Path() == "/health" }
	Skip func(c *fiber.Ctx) bool
}

// AccessLog returns a middleware with default configuration.
// You must provide a logger via AccessLogWithConfig if you want to use this.
//
// Example usage:
//
//	logger, _ := zap.NewProduction()
//	app.Use(middleware.AccessLogWithConfig(&middleware.AccessLogConfig{
//	    Logger: logger,
//	}))
func AccessLog() fiber.Handler {
	return AccessLogWithConfig(&AccessLogConfig{})
}

// AccessLogWithConfig allows customizing access log behaviour.
//
// Example usage:
//
//	logger, _ := zap.NewProduction()
//	app.Use(middleware.AccessLogWithConfig(&middleware.AccessLogConfig{
//	    Logger: logger,
//	    IncludeHeaders: []string{"X-Request-ID", "User-Agent"},
//	    Skip: func(c *fiber.Ctx) bool {
//	        return c.Path() == "/health" || c.Path() == "/metrics"
//	    },
//	}))
func AccessLogWithConfig(cfg *AccessLogConfig) fiber.Handler {
	// Set defaults
	if cfg.LevelResolver == nil {
		cfg.LevelResolver = defaultLevelResolver
	}

	return func(c *fiber.Ctx) error {
		// Skip if configured
		if cfg.Skip != nil && cfg.Skip(c) {
			return c.Next()
		}

		start := time.Now()
		err := c.Next()
		duration := time.Since(start)

		// Determine status code
		status := determineStatus(c, err)
		
		// Determine log level
		level := cfg.LevelResolver(status, err)

		// Build log fields
		fields := []zap.Field{
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", status),
			zap.Duration("duration", duration),
			zap.String("ip", c.IP()),
		}

		// Add configured headers
		for _, header := range cfg.IncludeHeaders {
			if val := c.Get(header); val != "" {
				fields = append(fields, zap.String("header_"+header, val))
			}
		}

		// Add error if present
		if err != nil {
			fields = append(fields, zap.Error(err))
		}

		// Log based on level
		if cfg.Logger != nil {
			switch level {
			case zapcore.DebugLevel:
				cfg.Logger.Debug("http request", fields...)
			case zapcore.InfoLevel:
				cfg.Logger.Info("http request", fields...)
			case zapcore.WarnLevel:
				cfg.Logger.Warn("http request", fields...)
			case zapcore.ErrorLevel:
				cfg.Logger.Error("http request", fields...)
			default:
				cfg.Logger.Info("http request", fields...)
			}
		}

		return err
	}
}

// defaultLevelResolver returns appropriate log level based on status code.
func defaultLevelResolver(status int, err error) zapcore.Level {
	switch {
	case status >= 500:
		return zapcore.ErrorLevel
	case status >= 400:
		return zapcore.WarnLevel
	case err != nil:
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// determineStatus extracts the response status code.
func determineStatus(c *fiber.Ctx, err error) int {
	if err != nil {
		if e, ok := err.(*fiber.Error); ok {
			return e.Code
		}
		return fiber.StatusInternalServerError
	}
	return c.Response().StatusCode()
}

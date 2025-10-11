package middleware

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// ErrorResponse is the standard error response structure.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// ErrorHandlerConfig defines configuration for the error handler.
type ErrorHandlerConfig struct {
	// Logger for logging internal errors (optional)
	Logger *zap.Logger
	
	// HideInternalErrors when true, returns generic message for non-Fiber errors (default: true)
	HideInternalErrors bool
}

// ErrorHandler returns a fiber error handler producing JSON responses.
//
// SECURITY: Internal errors are logged but NOT exposed to clients to prevent information disclosure.
//
// Error handling rules:
// - Fiber errors (*fiber.Error) are considered safe to expose
// - All other errors are logged and return generic "Internal Server Error"
//
// Example usage:
//
//	app := fiber.New(fiber.Config{
//	    ErrorHandler: middleware.ErrorHandler(),
//	})
func ErrorHandler() fiber.ErrorHandler {
	return ErrorHandlerWithConfig(ErrorHandlerConfig{
		HideInternalErrors: true,
	})
}

// ErrorHandlerWithConfig returns an error handler with custom configuration.
//
// Example usage:
//
//	logger, _ := zap.NewProduction()
//	app := fiber.New(fiber.Config{
//	    ErrorHandler: middleware.ErrorHandlerWithConfig(middleware.ErrorHandlerConfig{
//	        Logger:             logger,
//	        HideInternalErrors: true,
//	    }),
//	})
func ErrorHandlerWithConfig(cfg ErrorHandlerConfig) fiber.ErrorHandler {
	// Default to hiding internal errors
	if !cfg.HideInternalErrors {
		cfg.HideInternalErrors = true
	}

	return func(c *fiber.Ctx, err error) error {
		// Fiber errors are considered safe to expose (they're explicitly created by handlers)
		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			return c.Status(fiberErr.Code).JSON(ErrorResponse{
				Error:   fiberErr.Message,
				Message: fiberErr.Message,
			})
		}

		// SECURITY: Log internal errors for debugging but return generic message to client
		if cfg.Logger != nil {
			cfg.Logger.Error("internal error",
				zap.String("path", c.Path()),
				zap.String("method", c.Method()),
				zap.Error(err),
			)
		}

		// Return generic error message - do NOT expose internal error details
		if cfg.HideInternalErrors {
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
				Error:   "Internal Server Error",
				Message: "An unexpected error occurred",
			})
		}

		// If HideInternalErrors is false (not recommended for production)
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   err.Error(),
			Message: err.Error(),
		})
	}
}

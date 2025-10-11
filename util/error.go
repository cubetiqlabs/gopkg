package util

import (
	"github.com/gofiber/fiber/v2"
)

// HTTP status code constants
const (
	StatusConflict            = fiber.StatusConflict
	StatusNotFound            = fiber.StatusNotFound
	StatusBadRequest          = fiber.StatusBadRequest
	StatusInternalServerError = fiber.StatusInternalServerError
)

// NewError creates a new Fiber error with the given status code and message.
func NewError(status int, message string) error {
	return fiber.NewError(status, message)
}

// NotFoundErr wraps an error as a 404 Not Found error.
func NotFoundErr(err error) error {
	return fiber.NewError(StatusNotFound, err.Error())
}

// NotFoundError creates a 404 Not Found error with a custom message.
func NotFoundError(message string) error {
	return fiber.NewError(StatusNotFound, message)
}

// BadRequestErr wraps an error as a 400 Bad Request error.
func BadRequestErr(err error) error {
	return fiber.NewError(StatusBadRequest, err.Error())
}

// BadRequestError creates a 400 Bad Request error with a custom message.
func BadRequestError(message string) error {
	return fiber.NewError(StatusBadRequest, message)
}

// UnauthorizedError creates a 401 Unauthorized error with a custom message.
func UnauthorizedError(message string) error {
	return fiber.NewError(fiber.StatusUnauthorized, message)
}

// ForbiddenError creates a 403 Forbidden error with a custom message.
func ForbiddenError(message string) error {
	return fiber.NewError(fiber.StatusForbidden, message)
}

// InternalServerErr wraps an error as a 500 Internal Server Error.
func InternalServerErr(err error) error {
	return fiber.NewError(StatusInternalServerError, err.Error())
}

// InternalServerError creates a 500 Internal Server Error with a custom message.
func InternalServerError(message string) error {
	return fiber.NewError(StatusInternalServerError, message)
}

// ConflictErr wraps an error as a 409 Conflict error.
func ConflictErr(err error) error {
	return fiber.NewError(StatusConflict, err.Error())
}

// ConflictError creates a 409 Conflict error with a custom message.
func ConflictError(message string) error {
	return fiber.NewError(StatusConflict, message)
}

// UnprocessableEntityError creates a 422 Unprocessable Entity error with a custom message.
func UnprocessableEntityError(message string) error {
	return fiber.NewError(fiber.StatusUnprocessableEntity, message)
}

// NotImplementedError creates a 501 Not Implemented error with a custom message.
func NotImplementedError(message string) error {
	return fiber.NewError(fiber.StatusNotImplemented, message)
}

// TooManyRequestsError creates a 429 Too Many Requests error with a custom message.
func TooManyRequestsError(message string) error {
	return fiber.NewError(fiber.StatusTooManyRequests, message)
}

// ServiceUnavailableError creates a 503 Service Unavailable error with a custom message.
func ServiceUnavailableError(message string) error {
	return fiber.NewError(fiber.StatusServiceUnavailable, message)
}

package util

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestNotFoundError(t *testing.T) {
	err := NotFoundError("user not found")

	fiberErr, ok := err.(*fiber.Error)
	assert.True(t, ok, "should return fiber.Error")
	assert.Equal(t, fiber.StatusNotFound, fiberErr.Code)
	assert.Equal(t, "user not found", fiberErr.Message)
}

func TestBadRequestError(t *testing.T) {
	err := BadRequestError("invalid input")

	fiberErr, ok := err.(*fiber.Error)
	assert.True(t, ok, "should return fiber.Error")
	assert.Equal(t, fiber.StatusBadRequest, fiberErr.Code)
	assert.Equal(t, "invalid input", fiberErr.Message)
}

func TestUnauthorizedError(t *testing.T) {
	err := UnauthorizedError("invalid token")

	fiberErr, ok := err.(*fiber.Error)
	assert.True(t, ok, "should return fiber.Error")
	assert.Equal(t, fiber.StatusUnauthorized, fiberErr.Code)
	assert.Equal(t, "invalid token", fiberErr.Message)
}

func TestForbiddenError(t *testing.T) {
	err := ForbiddenError("access denied")

	fiberErr, ok := err.(*fiber.Error)
	assert.True(t, ok, "should return fiber.Error")
	assert.Equal(t, fiber.StatusForbidden, fiberErr.Code)
	assert.Equal(t, "access denied", fiberErr.Message)
}

func TestConflictError(t *testing.T) {
	err := ConflictError("resource already exists")

	fiberErr, ok := err.(*fiber.Error)
	assert.True(t, ok, "should return fiber.Error")
	assert.Equal(t, fiber.StatusConflict, fiberErr.Code)
	assert.Equal(t, "resource already exists", fiberErr.Message)
}

func TestInternalServerError(t *testing.T) {
	err := InternalServerError("database connection failed")

	fiberErr, ok := err.(*fiber.Error)
	assert.True(t, ok, "should return fiber.Error")
	assert.Equal(t, fiber.StatusInternalServerError, fiberErr.Code)
	assert.Equal(t, "database connection failed", fiberErr.Message)
}

func TestServiceUnavailableError(t *testing.T) {
	err := ServiceUnavailableError("service temporarily unavailable")

	fiberErr, ok := err.(*fiber.Error)
	assert.True(t, ok, "should return fiber.Error")
	assert.Equal(t, fiber.StatusServiceUnavailable, fiberErr.Code)
	assert.Equal(t, "service temporarily unavailable", fiberErr.Message)
}

func TestTooManyRequestsError(t *testing.T) {
	err := TooManyRequestsError("rate limit exceeded")

	fiberErr, ok := err.(*fiber.Error)
	assert.True(t, ok, "should return fiber.Error")
	assert.Equal(t, fiber.StatusTooManyRequests, fiberErr.Code)
	assert.Equal(t, "rate limit exceeded", fiberErr.Message)
}

func TestNewError(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		message  string
		expected int
	}{
		{"Custom 418", 418, "I'm a teapot", 418},
		{"Custom 503", 503, "service down", 503},
		{"Custom 422", 422, "unprocessable entity", 422},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewError(tt.code, tt.message)

			fiberErr, ok := err.(*fiber.Error)
			assert.True(t, ok, "should return fiber.Error")
			assert.Equal(t, tt.expected, fiberErr.Code)
			assert.Equal(t, tt.message, fiberErr.Message)
		})
	}
}

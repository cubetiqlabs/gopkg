package logging

import (
	"context"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
	once   sync.Once
)

// Init initializes a global zap logger. Safe to call multiple times; first call wins.
//
// Parameters:
//   - level: Log level (debug, info, warn, error, dpanic, panic, fatal)
//   - development: If true, enables development mode with stack traces and DPanic
//
// Example usage:
//
//	logger, err := logging.Init("info", false)
//	if err != nil {
//	    panic(err)
//	}
//	defer logger.Sync()
func Init(level string, development bool) (*zap.Logger, error) {
	var err error
	var stackKey string
	if development {
		stackKey = "stack"
	}

	once.Do(func() {
		cfg := zap.Config{
			Level:       zap.NewAtomicLevelAt(parseLevel(level)),
			Development: development,
			Encoding:    "json",
			EncoderConfig: zapcore.EncoderConfig{
				TimeKey:       "ts",
				LevelKey:      "level",
				NameKey:       "logger",
				CallerKey:     "caller",
				MessageKey:    "msg",
				StacktraceKey: stackKey,
				EncodeTime:    zapcore.ISO8601TimeEncoder,
				EncodeLevel:   zapcore.LowercaseLevelEncoder,
				EncodeCaller:  zapcore.ShortCallerEncoder,
			},
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
		}
		logger, err = cfg.Build()
	})

	return logger, err
}

// parseLevel converts a string level to zapcore.Level.
func parseLevel(lvl string) zapcore.Level {
	switch lvl {
	case "debug":
		return zapcore.DebugLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// L returns the global logger. Panics if not initialized.
// Use Init() before calling this function.
func L() *zap.Logger {
	if logger == nil {
		panic("logger not initialized, call logging.Init() first")
	}
	return logger
}

// WithContext stores logger with fields inside context.
// This is useful for adding request-scoped fields to logs.
//
// Example:
//
//	ctx := logging.WithContext(ctx, zap.String("request_id", rid))
//	logging.FromContext(ctx).Info("processing request")
func WithContext(ctx context.Context, fields ...zap.Field) context.Context {
	return context.WithValue(ctx, ctxKeyLogger{}, L().With(fields...))
}

// FromContext extracts logger from context or returns global logger.
// This allows request-scoped logging without passing logger explicitly.
//
// Example:
//
//	func handleRequest(ctx context.Context) {
//	    logging.FromContext(ctx).Info("handling request")
//	}
func FromContext(ctx context.Context) *zap.Logger {
	if v := ctx.Value(ctxKeyLogger{}); v != nil {
		if lg, ok := v.(*zap.Logger); ok {
			return lg
		}
	}
	return L()
}

type ctxKeyLogger struct{}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Info(msg, fields...)
	}
}

// Infof logs an info message with format
func Infof(msg string, args ...interface{}) {
	if logger != nil {
		logger.Sugar().Infof(msg, args...)
	}
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Debug(msg, fields...)
	}
}

// Debugf logs a debug message with format
func Debugf(msg string, args ...interface{}) {
	if logger != nil {
		logger.Sugar().Debugf(msg, args...)
	}
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Warn(msg, fields...)
	}
}

// Warnf logs a warning message with format
func Warnf(msg string, args ...interface{}) {
	if logger != nil {
		logger.Sugar().Warnf(msg, args...)
	}
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Error(msg, fields...)
	}
}

// Errorf logs an error message with format
func Errorf(msg string, args ...interface{}) {
	if logger != nil {
		logger.Sugar().Errorf(msg, args...)
	}
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Fatal(msg, fields...)
	}
}

// Fatalf logs a fatal message with format and exits
func Fatalf(msg string, args ...interface{}) {
	if logger != nil {
		logger.Sugar().Fatalf(msg, args...)
	}
}

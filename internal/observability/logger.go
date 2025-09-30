package observability

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

// LogLevel represents the logging level configuration
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

// Config holds the configuration for the logger
type Config struct {
	Level     LogLevel
	Output    io.Writer
	AddSource bool
}

// DefaultConfig returns the default logger configuration
func DefaultConfig() Config {
	return Config{
		Level:     LevelInfo,
		Output:    os.Stdout,
		AddSource: false,
	}
}

// NewLogger creates a new structured logger with the given configuration
func NewLogger(cfg Config) *slog.Logger {
	level := slogLevel(cfg.Level)
	
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize timestamp format
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(a.Value.Time().UTC().Format("2006-01-02T15:04:05.000Z"))
			}
			return a
		},
	}
	
	var handler slog.Handler
	
	// Use JSON handler for production-like environments, text for development
	env := strings.ToLower(os.Getenv("OMNIDROP_ENV"))
	if env == "production" || env == "staging" {
		handler = slog.NewJSONHandler(cfg.Output, opts)
	} else {
		handler = slog.NewTextHandler(cfg.Output, opts)
	}
	
	return slog.New(handler)
}

// SetupLogger configures the global slog logger based on environment
func SetupLogger() *slog.Logger {
	cfg := DefaultConfig()
	
	// Configure based on environment
	env := strings.ToLower(os.Getenv("OMNIDROP_ENV"))
	switch env {
	case "development", "dev":
		cfg.Level = LevelDebug
		cfg.AddSource = true
	case "test", "testing":
		cfg.Level = LevelWarn
		cfg.Output = io.Discard // Suppress logs during testing
	case "production", "prod":
		cfg.Level = LevelInfo
		cfg.AddSource = false
	default:
		// Default to info level
		cfg.Level = LevelInfo
	}
	
	// Allow override via LOG_LEVEL environment variable
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		switch strings.ToLower(logLevel) {
		case "debug":
			cfg.Level = LevelDebug
		case "info":
			cfg.Level = LevelInfo
		case "warn", "warning":
			cfg.Level = LevelWarn
		case "error":
			cfg.Level = LevelError
		}
	}
	
	logger := NewLogger(cfg)
	
	// Set as default logger
	slog.SetDefault(logger)
	
	return logger
}

// slogLevel converts our LogLevel to slog.Level
func slogLevel(level LogLevel) slog.Level {
	switch level {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WithRequestID adds a request ID to the logger context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	logger := slog.With("request_id", requestID)
	return context.WithValue(ctx, "logger", logger)
}

// LoggerFromContext retrieves the logger from context, falling back to default
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value("logger").(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}
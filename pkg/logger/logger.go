package logger

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	// Logger is the global logger instance
	Logger zerolog.Logger
	// MetricsHook can be used to integrate with metrics/observability systems like Prometheus
	MetricsHook zerolog.Hook
)

// Init initializes the global logger with the specified log level
func Init(logLevel string) {
	initWithWriter(os.Stdout, logLevel)
}

// InitWithWriter initializes the logger with a custom writer
func InitWithWriter(w io.Writer, logLevel string) {
	initWithWriter(w, logLevel)
}

func initWithWriter(w io.Writer, logLevel string) {
	// Parse log level
	level := parseLogLevel(logLevel)
	zerolog.SetGlobalLevel(level)

	// Configure human-readable console output
	output := zerolog.ConsoleWriter{
		Out:        w,
		TimeFormat: time.RFC3339,
		NoColor:    false,
	}

	// Create logger with optional metrics hook for future observability integration
	if MetricsHook != nil {
		Logger = zerolog.New(output).
			With().
			Timestamp().
			Logger().
			Hook(MetricsHook)
	} else {
		Logger = zerolog.New(output).
			With().
			Timestamp().
			Logger()
	}

	// Set as global logger
	log.Logger = Logger
}

// parseLogLevel converts string log level to zerolog.Level
func parseLogLevel(level string) zerolog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

// GetLogger returns the global logger instance
func GetLogger() *zerolog.Logger {
	return &Logger
}

// SetMetricsHook sets a hook for metrics/observability integration
// This can be used to integrate with systems like Prometheus in the future
func SetMetricsHook(hook zerolog.Hook) {
	MetricsHook = hook
}

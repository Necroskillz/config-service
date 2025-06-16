package logging

import (
	"log/slog"
	"os"
)

func ConfigureSlog(app string) *slog.Logger {
	// Configure logging based on environment
	var logHandler slog.Handler

	isDevelopment := os.Getenv("ENV") == "dev"
	logLevel := slog.LevelInfo

	switch os.Getenv("LOG_LEVEL") {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	if isDevelopment {
		// Human-readable format for development
		logHandler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: logLevel,
		})
	} else {
		// JSON format for production (better for log aggregation)
		logHandler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level:     logLevel,
			AddSource: true,
		})
	}

	logger := slog.New(logHandler).With("app", app)
	slog.SetDefault(logger)

	return logger
}

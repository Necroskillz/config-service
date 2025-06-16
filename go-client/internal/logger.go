package internal

import (
	"context"
	"log/slog"
)

type Logger interface {
	Error(ctx context.Context, msg string, fields ...any)
	Info(ctx context.Context, msg string, fields ...any)
	Warn(ctx context.Context, msg string, fields ...any)
	Debug(ctx context.Context, msg string, fields ...any)
}

type LoggerImpl struct {
	logFunc func(ctx context.Context, level slog.Level, msg string, fields ...any)
}

func NewLogger(logFunc func(ctx context.Context, level slog.Level, msg string, fields ...any)) Logger {
	if logFunc == nil {
		logFunc = func(ctx context.Context, level slog.Level, msg string, fields ...any) {
		}
	}

	return &LoggerImpl{
		logFunc: logFunc,
	}
}

func (l *LoggerImpl) Error(ctx context.Context, msg string, fields ...any) {
	l.logFunc(ctx, slog.LevelError, msg, fields...)
}

func (l *LoggerImpl) Info(ctx context.Context, msg string, fields ...any) {
	l.logFunc(ctx, slog.LevelInfo, msg, fields...)
}

func (l *LoggerImpl) Warn(ctx context.Context, msg string, fields ...any) {
	l.logFunc(ctx, slog.LevelWarn, msg, fields...)
}

func (l *LoggerImpl) Debug(ctx context.Context, msg string, fields ...any) {
	l.logFunc(ctx, slog.LevelDebug, msg, fields...)
}

package logger

import (
	"context"
	"log/slog"
	"os"
)

type contextKey string

const loggerKey contextKey = "logger"

func New(env string) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: env == "production",
	}
	var handler slog.Handler
	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	l := slog.New(handler)
	slog.SetDefault(l)
	return l
}

func WithContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}

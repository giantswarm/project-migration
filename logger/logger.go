package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"
)

// Logger is the global logger instance.
var Logger *slog.Logger

func init() {
	Logger = slog.New(NewColoredHandler(os.Stdout))
}

// NewColoredHandler returns a new slog.Handler that outputs colored logs.
func NewColoredHandler(out *os.File) slog.Handler {
	return &coloredHandler{out: out}
}

type coloredHandler struct {
	out *os.File
}

func (h *coloredHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *coloredHandler) Handle(ctx context.Context, r slog.Record) error {
	timestamp := time.Now().Format(time.RFC3339)
	color := levelToColor(r.Level)
	reset := "\033[0m"
	// Format: timestamp level message
	line := fmt.Sprintf("%s%s %-5s%s %s\n", color, timestamp, r.Level.String(), reset, r.Message)
	_, err := h.out.WriteString(line)
	return err
}

func (h *coloredHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *coloredHandler) WithGroup(name string) slog.Handler {
	return h
}

func levelToColor(l slog.Level) string {
	switch l {
	case slog.LevelDebug:
		return "\033[36m" // cyan
	case slog.LevelInfo:
		return "\033[32m" // green
	case slog.LevelWarn:
		return "\033[33m" // yellow
	case slog.LevelError:
		return "\033[31m" // red
	default:
		return "\033[0m"
	}
}

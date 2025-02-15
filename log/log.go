package log

import (
	"context"
	"fmt"
	"strings"

	"log/slog"
)

// CustomHandler implements slog.Handler for friendly colored output.
type CustomHandler struct{}

func (h *CustomHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *CustomHandler) Handle(ctx context.Context, r slog.Record) error {
	var color string
	if r.Level >= slog.LevelError {
		color = "\033[31m" // red for errors
	} else {
		color = "\033[36m" // cyan for info
	}
	var b strings.Builder
	b.WriteString(color)
	b.WriteString(r.Message)
	if r.NumAttrs() > 0 {
		b.WriteString(" (")
		first := true
		r.Attrs(func(a slog.Attr) bool {
			if !first {
				b.WriteString(", ")
			}
			first = false
			b.WriteString(a.Key)
			b.WriteString("=")
			b.WriteString(fmt.Sprint(a.Value.Any()))
			return true
		})
		b.WriteString(")")
	}
	b.WriteString("\033[0m\n")
	fmt.Print(b.String())
	return nil
}

func (h *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *CustomHandler) WithGroup(name string) slog.Handler {
	return h
}

// NewHandler creates a new instance of CustomHandler.
func NewHandler() slog.Handler {
	return &CustomHandler{}
}

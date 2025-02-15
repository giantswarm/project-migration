package log

import (
	"bytes"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"
)

func captureOutput(f func()) string {
	// temporarily replace stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = old
	return buf.String()
}

func TestCustomHandlerInfoOutput(t *testing.T) {
	handler := NewHandler()
	// Create a logger with our custom handler.
	logger := slog.New(handler)

	// Log an info message with an attribute.
	output := captureOutput(func() {
		logger.Info("Test message", slog.String("key", "value"))
		// Give time for asynchronous writes.
		time.Sleep(10 * time.Millisecond)
	})

	// Expect cyan color for info.
	if !strings.Contains(output, "\033[36m") {
		t.Errorf("Expected cyan color code \\033[36m in output, got: %s", output)
	}
	// Expect message and attribute.
	if !strings.Contains(output, "Test message") ||
		!strings.Contains(output, "key=value") ||
		!strings.Contains(output, "\033[0m") {
		t.Errorf("Output formatting incorrect: %s", output)
	}
}

func TestCustomHandlerErrorOutput(t *testing.T) {
	handler := NewHandler()
	logger := slog.New(handler)

	output := captureOutput(func() {
		logger.Error("Error occurred", slog.Int("code", 500))
		time.Sleep(10 * time.Millisecond)
	})

	// Expect red color for errors.
	if !strings.Contains(output, "\033[31m") {
		t.Errorf("Expected red color code \\033[31m in output, got: %s", output)
	}
	// Verify message and attribute.
	if !strings.Contains(output, "Error occurred") ||
		!strings.Contains(output, "code=500") ||
		!strings.Contains(output, "\033[0m") {
		t.Errorf("Output formatting incorrect: %s", output)
	}
}

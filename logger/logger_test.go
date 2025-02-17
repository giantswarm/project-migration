package logger

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewColoredHandler(t *testing.T) {
	// Use os.Pipe to capture output.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	handler := NewColoredHandler(w)

	// Build a minimal record.
	record := slog.Record{
		Time:    time.Now(),
		Level:   slog.LevelInfo,
		Message: "Test message",
	}
	if err := handler.Handle(context.Background(), record); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	// Close writer to signal EOF.
	w.Close()

	// Read output from the pipe.
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(r)
	if err != nil {
		t.Fatalf("Error reading from pipe: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "Test message") {
		t.Errorf("Expected output to contain 'Test message', got: %s", output)
	}
}

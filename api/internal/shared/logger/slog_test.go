package logger_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/paroki/domus/api/internal/shared/logger"
)

func TestSlogLogger(t *testing.T) {
	var buf bytes.Buffer
	h := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	sl := logger.NewSlogLogger(slog.New(h))

	// Test Info
	sl.Info("test info message", "key1", "val1")

	var data map[string]any
	if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
		t.Fatalf("failed to unmarshal JSON log: %v", err)
	}

	if data["msg"] != "test info message" {
		t.Errorf("expected msg to be 'test info message', got %v", data["msg"])
	}
	if data["level"] != "INFO" {
		t.Errorf("expected level to be 'INFO', got %v", data["level"])
	}
	if data["key1"] != "val1" {
		t.Errorf("expected key1 to be 'val1', got %v", data["key1"])
	}

	// Reset buffer
	buf.Reset()

	// Test With
	withLogger := sl.With("ctxKey", "ctxVal")
	withLogger.Debug("test debug message with context", "localKey", "localVal")

	data = make(map[string]any)
	if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
		t.Fatalf("failed to unmarshal JSON log: %v", err)
	}

	if data["msg"] != "test debug message with context" {
		t.Errorf("expected msg to be 'test debug message with context', got %v", data["msg"])
	}
	if data["level"] != "DEBUG" {
		t.Errorf("expected level to be 'DEBUG', got %v", data["level"])
	}
	if data["ctxKey"] != "ctxVal" {
		t.Errorf("expected ctxKey to be 'ctxVal', got %v", data["ctxKey"])
	}
	if data["localKey"] != "localVal" {
		t.Errorf("expected localKey to be 'localVal', got %v", data["localKey"])
	}
}

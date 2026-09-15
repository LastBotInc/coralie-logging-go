package clog

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel/trace"
)

func TestConsoleSink(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Console.Enabled = true
	cfg.Console.Colors = false // Disable colors for test
	Init(cfg)
	defer Shutdown(context.Background())

	Info("Test", "Console message")
	time.Sleep(100 * time.Millisecond)
}

func TestFileSink(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Console.Enabled = false
	cfg.File.BaseDir = tmpDir
	cfg.File.PerLevel = map[Level]string{
		LevelInfo:  "info.log",
		LevelError: "error.log",
	}

	Init(cfg)
	defer Shutdown(context.Background())

	Info("Test", "Info message")
	Error("Test", "Error message")
	Debug("Test", "Debug message") // Should not be written

	time.Sleep(200 * time.Millisecond)

	// Verify files exist and contain expected content
	infoFile := filepath.Join(tmpDir, "info.log")
	errorFile := filepath.Join(tmpDir, "error.log")

	infoContent, err := os.ReadFile(infoFile)
	if err != nil {
		t.Fatalf("Failed to read info.log: %v", err)
	}
	if len(infoContent) == 0 {
		t.Error("info.log is empty")
	}

	errorContent, err := os.ReadFile(errorFile)
	if err != nil {
		t.Fatalf("Failed to read error.log: %v", err)
	}
	if len(errorContent) == 0 {
		t.Error("error.log is empty")
	}

	// Verify debug was not written
	debugFile := filepath.Join(tmpDir, "debug.log")
	if _, err := os.Stat(debugFile); err == nil {
		t.Error("debug.log should not exist")
	}
}

func TestFileRouting(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Console.Enabled = false
	cfg.File.BaseDir = tmpDir
	cfg.File.PerLevel = map[Level]string{
		LevelError: "errors.log",
	}

	Init(cfg)
	defer Shutdown(context.Background())

	Info("Test", "Info message") // Should not be written to file
	Error("Test", "Error message")

	time.Sleep(200 * time.Millisecond)

	errorFile := filepath.Join(tmpDir, "errors.log")
	errorContent, err := os.ReadFile(errorFile)
	if err != nil {
		t.Fatalf("Failed to read errors.log: %v", err)
	}
	if len(errorContent) == 0 {
		t.Error("errors.log is empty")
	}
}

func TestOmitLevel(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Console.Enabled = true
	cfg.Console.OmitLevels = map[Level]bool{
		LevelDebug: true,
	}
	Init(cfg)
	defer Shutdown(context.Background())

	Debug("Test", "Should be omitted")
	Info("Test", "Should be shown")

	time.Sleep(100 * time.Millisecond)
}

func TestContextfulConsoleAndFileCorrelation(t *testing.T) {
	dir := t.TempDir()
	consolePath := filepath.Join(dir, "console.log")
	console, err := os.Create(consolePath)
	if err != nil {
		t.Fatal(err)
	}
	originalStdout := os.Stdout
	os.Stdout = console
	t.Cleanup(func() {
		Shutdown(context.Background())
		os.Stdout = originalStdout
		if err := console.Close(); err != nil {
			t.Error(err)
		}
	})
	previousRedaction := RedactionEnabled()
	SetRedactionEnabled(true)
	t.Cleanup(func() { SetRedactionEnabled(previousRedaction) })
	cfg := DefaultConfig()
	cfg.Console.Enabled = true
	cfg.Console.Colors = false
	cfg.Dedupe.Enabled = false
	cfg.File.BaseDir = dir
	cfg.File.PerLevel = map[Level]string{LevelInfo: "events.log"}
	Init(cfg)
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{1}, SpanID: trace.SpanID{2},
	})
	ctx := trace.ContextWithSpanContext(context.Background(), spanContext)
	LogContext(ctx, LevelInfo, "Call", "completed for %s@%s", "synthetic", "example.invalid")
	Info("Call", "legacy message")
	LogContext(context.Background(), LevelInfo, "Call", "background message")
	Shutdown(context.Background())

	for _, path := range []string{consolePath, filepath.Join(dir, "events.log")} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		if len(lines) != 3 {
			t.Fatalf("%s: got %d lines, want 3", filepath.Base(path), len(lines))
		}
		wantCorrelation := "[trace_id=" + spanContext.TraceID().String() + " span_id=" + spanContext.SpanID().String() + "] "
		if !strings.Contains(lines[0], "[INFO][Call]"+wantCorrelation+"completed for ") {
			t.Errorf("%s: contextful record lost correlation", filepath.Base(path))
		}
		if strings.Contains(lines[0], "synthetic@example.invalid") {
			t.Errorf("%s: contextful record bypassed redaction", filepath.Base(path))
		}
		for i, message := range []string{"legacy message", "background message"} {
			if !strings.HasSuffix(lines[i+1], "[INFO][Call]"+message) || strings.Contains(lines[i+1], "trace_id=") || strings.Contains(lines[i+1], "span_id=") {
				t.Errorf("%s: contextless record changed", filepath.Base(path))
			}
		}
	}
}

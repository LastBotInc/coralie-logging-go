package clog

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
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


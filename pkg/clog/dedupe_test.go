package clog

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDedupe_CollapsesConsecutive(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Console.Enabled = false
	cfg.File.BaseDir = tmpDir
	cfg.File.PerLevel = map[Level]string{
		LevelInfo: "test.log",
	}
	cfg.Dedupe.Enabled = true

	Init(cfg)
	defer Shutdown(context.Background())

	// Log same message multiple times
	for i := 0; i < 5; i++ {
		Info("Test", "Same message")
	}

	// Log different message
	Info("Test", "Different message")

	time.Sleep(500 * time.Millisecond)

	// Read log file
	logFile := filepath.Join(tmpDir, "test.log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	contentStr := string(content)
	
	// Should have the first message
	if !strings.Contains(contentStr, "Same message") {
		t.Error("Log should contain 'Same message'")
	}

	// Should have summary
	if !strings.Contains(contentStr, "repeated") {
		t.Errorf("Log should contain deduplication summary. Content: %q", contentStr)
	}

	// Should have different message
	if !strings.Contains(contentStr, "Different message") {
		t.Error("Log should contain 'Different message'")
	}

	// Should not have 5 instances of "Same message"
	count := strings.Count(contentStr, "Same message")
	if count > 2 { // One original + maybe one in summary context
		t.Errorf("Expected at most 2 instances of 'Same message', got %d", count)
	}
}

func TestDedupe_DoesNotCollapseNonConsecutive(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Console.Enabled = false
	cfg.File.BaseDir = tmpDir
	cfg.File.PerLevel = map[Level]string{
		LevelInfo: "test.log",
	}
	cfg.Dedupe.Enabled = true

	Init(cfg)
	defer Shutdown(context.Background())

	Info("Test", "Message A")
	Info("Test", "Message B")
	Info("Test", "Message A") // Non-consecutive duplicate

	time.Sleep(300 * time.Millisecond)

	logFile := filepath.Join(tmpDir, "test.log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	contentStr := string(content)
	
	// Both instances of Message A should appear (non-consecutive)
	count := strings.Count(contentStr, "Message A")
	if count < 2 {
		t.Errorf("Expected at least 2 instances of 'Message A' (non-consecutive), got %d", count)
	}
}

func TestDedupe_LevelOrIfaceBreaks(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Console.Enabled = false
	cfg.File.BaseDir = tmpDir
	cfg.File.PerLevel = map[Level]string{
		LevelInfo: "test.log",
	}
	cfg.Dedupe.Enabled = true

	Init(cfg)
	defer Shutdown(context.Background())

	Info("Test", "Same message")
	Info("Test", "Same message")
	Error("Test", "Same message") // Different level - should break dedupe
	Info("Test", "Same message")
	Info("Other", "Same message") // Different iface - should break dedupe

	time.Sleep(300 * time.Millisecond)

	logFile := filepath.Join(tmpDir, "test.log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	contentStr := string(content)
	
	// Should have multiple instances due to level/iface breaks
	count := strings.Count(contentStr, "Same message")
	if count < 3 {
		t.Errorf("Expected at least 3 instances of 'Same message' (level/iface breaks), got %d", count)
	}
}

func TestDedupe_FlushOnShutdown(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Console.Enabled = false
	cfg.File.BaseDir = tmpDir
	cfg.File.PerLevel = map[Level]string{
		LevelInfo: "test.log",
	}
	cfg.Dedupe.Enabled = true

	Init(cfg)

	// Log same message multiple times
	for i := 0; i < 3; i++ {
		Info("Test", "Pending message")
	}

	// Shutdown immediately (should flush summary)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	Shutdown(ctx)

	time.Sleep(100 * time.Millisecond)

	logFile := filepath.Join(tmpDir, "test.log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	contentStr := string(content)
	
	// Should have summary flushed
	if !strings.Contains(contentStr, "repeated") {
		t.Error("Log should contain deduplication summary after shutdown")
	}
}

func TestDedupe_Disabled(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Console.Enabled = false
	cfg.File.BaseDir = tmpDir
	cfg.File.PerLevel = map[Level]string{
		LevelInfo: "test.log",
	}
	cfg.Dedupe.Enabled = false

	Init(cfg)
	defer Shutdown(context.Background())

	// Log same message multiple times
	for i := 0; i < 5; i++ {
		Info("Test", "Same message")
	}

	time.Sleep(300 * time.Millisecond)

	logFile := filepath.Join(tmpDir, "test.log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	contentStr := string(content)
	
	// Should have all 5 instances when dedupe is disabled
	count := strings.Count(contentStr, "Same message")
	if count < 5 {
		t.Errorf("Expected 5 instances of 'Same message' when dedupe disabled, got %d", count)
	}
}


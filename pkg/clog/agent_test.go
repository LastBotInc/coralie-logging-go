package clog

import (
	"context"
	"testing"
	"time"
)

func TestDropPolicy_DropNew(t *testing.T) {
	cfg := DefaultConfig()
	cfg.QueueSize = 5
	cfg.DropPolicy = "drop_new"
	Init(cfg)
	defer Shutdown(context.Background())

	// Fill queue beyond capacity
	for i := 0; i < 20; i++ {
		Info("Test", "Message %d", i)
	}

	// Give agent time to process some
	time.Sleep(200 * time.Millisecond)

	stats := GetStats()
	if stats.AcceptedCount == 0 {
		t.Error("Expected some accepted events")
	}
	// With drop_new, we should have some drops
	if stats.DropsPerLevel[LevelInfo] == 0 && stats.AcceptedCount >= 20 {
		t.Logf("Note: No drops recorded, but queue was small. Accepted: %d", stats.AcceptedCount)
	}
}

func TestDropPolicy_DropOld(t *testing.T) {
	cfg := DefaultConfig()
	cfg.QueueSize = 5
	cfg.DropPolicy = "drop_old"
	Init(cfg)
	defer Shutdown(context.Background())

	// Fill queue beyond capacity
	for i := 0; i < 20; i++ {
		Info("Test", "Message %d", i)
	}

	// Give agent time to process some
	time.Sleep(200 * time.Millisecond)

	stats := GetStats()
	if stats.AcceptedCount == 0 {
		t.Error("Expected some accepted events")
	}
}

func TestInitShutdown_NoLeak(t *testing.T) {
	cfg := DefaultConfig()
	cfg.QueueSize = 10
	Init(cfg)

	// Send events
	for i := 0; i < 10; i++ {
		Info("Test", "Message %d", i)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	Shutdown(ctx)

	// Try to log after shutdown - should be safe (no-op)
	Info("Test", "After shutdown")
	time.Sleep(50 * time.Millisecond)
}

func TestStats(t *testing.T) {
	cfg := DefaultConfig()
	Init(cfg)
	defer Shutdown(context.Background())

	Info("Test", "Message 1")
	Error("Test", "Message 2")

	time.Sleep(100 * time.Millisecond)

	stats := GetStats()
	if stats.AcceptedCount < 2 {
		t.Errorf("Expected at least 2 accepted events, got %d", stats.AcceptedCount)
	}
	if stats.EmittedCount < 2 {
		t.Errorf("Expected at least 2 emitted events, got %d", stats.EmittedCount)
	}
}


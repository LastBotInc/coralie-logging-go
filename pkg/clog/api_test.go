package clog

import (
	"context"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	cfg := DefaultConfig()
	Init(cfg)
	defer Shutdown(context.Background())

	if !IsInitialized() {
		t.Error("IsInitialized() = false after Init()")
	}
}

func TestInit_MultipleCalls(t *testing.T) {
	cfg1 := DefaultConfig()
	cfg1.QueueSize = 100
	Init(cfg1)

	cfg2 := DefaultConfig()
	cfg2.QueueSize = 200
	Init(cfg2) // Should be ignored

	// First config should still be active
	if !IsInitialized() {
		t.Error("IsInitialized() = false")
	}

	Shutdown(context.Background())
}

func TestShutdown_DrainsQueue(t *testing.T) {
	cfg := DefaultConfig()
	cfg.QueueSize = 10
	Init(cfg)

	// Send some events
	for i := 0; i < 5; i++ {
		Info("Test", "Message %d", i)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	Shutdown(ctx)

	// Verify shutdown completed
	if IsInitialized() {
		t.Error("IsInitialized() = true after Shutdown()")
	}
}

func TestLogLevels(t *testing.T) {
	cfg := DefaultConfig()
	Init(cfg)
	defer Shutdown(context.Background())

	Debug("Test", "Debug message")
	Info("Test", "Info message")
	Success("Test", "Success message")
	Warning("Test", "Warning message")
	Fail("Test", "Fail message")
	Error("Test", "Error message")
	Catastrophe("Test", "Catastrophe message")

	// Give agent time to process
	time.Sleep(100 * time.Millisecond)
}

func TestDefaultIface(t *testing.T) {
	cfg := DefaultConfig()
	Init(cfg)
	defer Shutdown(context.Background())

	// Log without iface (empty string)
	Info("", "Message without iface")

	// Give agent time to process
	time.Sleep(100 * time.Millisecond)
}


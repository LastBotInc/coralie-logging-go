package clog

import (
	"context"
	"sync"
	"testing"
	"time"
)

// testHook is a test hook that records events.
type testHook struct {
	mu      sync.Mutex
	events  []Event
	callCount int
}

func (h *testHook) OnLog(e Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.events = append(h.events, e)
	h.callCount++
}

func (h *testHook) getEvents() []Event {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]Event(nil), h.events...)
}

func (h *testHook) getCallCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.callCount
}

func TestHookCalled(t *testing.T) {
	hook := &testHook{}

	cfg := DefaultConfig()
	cfg.Console.Enabled = false
	cfg.Hooks.Global = []Hook{hook}

	Init(cfg)
	defer Shutdown(context.Background())

	Info("Test", "Message 1")
	Error("Test", "Message 2")

	time.Sleep(200 * time.Millisecond)

	if hook.getCallCount() != 2 {
		t.Errorf("Expected hook to be called 2 times, got %d", hook.getCallCount())
	}

	events := hook.getEvents()
	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}
	if events[0].Level != LevelInfo {
		t.Errorf("Expected first event level to be Info, got %v", events[0].Level)
	}
	if events[1].Level != LevelError {
		t.Errorf("Expected second event level to be Error, got %v", events[1].Level)
	}
}

func TestPerLevelHook(t *testing.T) {
	globalHook := &testHook{}
	errorHook := &testHook{}

	cfg := DefaultConfig()
	cfg.Console.Enabled = false
	cfg.Hooks.Global = []Hook{globalHook}
	cfg.Hooks.PerLevel = map[Level][]Hook{
		LevelError: {errorHook},
	}

	Init(cfg)
	defer Shutdown(context.Background())

	Info("Test", "Info message")
	Error("Test", "Error message")
	Warning("Test", "Warning message")

	time.Sleep(200 * time.Millisecond)

	// Global hook should be called 3 times
	if globalHook.getCallCount() != 3 {
		t.Errorf("Expected global hook to be called 3 times, got %d", globalHook.getCallCount())
	}

	// Error hook should be called only once (for Error level)
	if errorHook.getCallCount() != 1 {
		t.Errorf("Expected error hook to be called 1 time, got %d", errorHook.getCallCount())
	}
}


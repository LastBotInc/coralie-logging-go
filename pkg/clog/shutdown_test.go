package clog

import (
	"context"
	"testing"
	"time"
)

func TestShutdown_DrainsAndFlushes(t *testing.T) {
	cfg := DefaultConfig()
	cfg.QueueSize = 10
	Init(cfg)

	// Send events
	for i := 0; i < 10; i++ {
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

func TestRecoverAndFlush_RepanicTrue(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Console.Enabled = false
	Init(cfg)
	defer Shutdown(context.Background())

	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		defer RecoverAndFlush(true)

		panic("test panic")
	}()

	if !panicked {
		t.Error("Expected panic to be re-raised")
	}
}

func TestRecoverAndFlush_RepanicFalse(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Console.Enabled = false
	Init(cfg)
	defer Shutdown(context.Background())

	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		defer RecoverAndFlush(false)

		panic("test panic")
	}()

	if panicked {
		t.Error("Expected panic to be suppressed when repanic=false")
	}
}

func TestInstallSignalHandler(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Console.Enabled = false
	Init(cfg)

	stop := InstallSignalHandler()

	// Just verify the handler can be installed and stopped
	// Don't actually send signals in tests as it causes os.Exit
	time.Sleep(10 * time.Millisecond)

	// Stop the handler
	stop()

	// Clean shutdown
	Shutdown(context.Background())
}

// TestSignalHandlerIntegration is skipped - signal handler calls os.Exit
// which is not testable in unit tests


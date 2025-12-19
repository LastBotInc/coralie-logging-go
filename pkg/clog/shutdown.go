// Package clog: shutdown, signal handling, and panic recovery.
package clog

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

// InstallSignalHandler installs signal handlers for SIGINT and SIGTERM.
// Returns a stop function that can be called to remove the handler.
func InstallSignalHandler() func() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	stop := make(chan struct{})
	stopped := make(chan struct{})

	go func() {
		defer close(stopped)
		select {
		case <-sigChan:
			// Signal received, shutdown with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			Shutdown(ctx)
			os.Exit(0)
		case <-stop:
			// Stop requested
			signal.Stop(sigChan)
		}
	}()

	return func() {
		close(stop)
		<-stopped
	}
}

// RecoverAndFlush recovers from panic and flushes logs.
// If repanic is true, it will re-panic after flushing.
// This is intended to be used with defer.
func RecoverAndFlush(repanic bool) {
	if r := recover(); r != nil {
		// Log the panic
		buf := make([]byte, 4096)
		n := runtime.Stack(buf, false)
		stackTrace := string(buf[:n])

		Catastrophe("System", "Panic recovered: %v\n%s", r, stackTrace)

		// Flush logs
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		Shutdown(ctx)

		// Re-panic if requested
		if repanic {
			panic(r)
		}
	}
}

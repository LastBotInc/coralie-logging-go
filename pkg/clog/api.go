// Package clog: public API functions for logging.
package clog

import (
	"context"
	"fmt"
	"sync"
)

const defaultIface = "Application"

var (
	globalAgent *agent
	initOnce    sync.Once
	initMu      sync.RWMutex
	shutdownMu  sync.Mutex
	shutdown    bool
)

// Init initializes the logger with the given configuration.
// It is safe to call multiple times, but only the first call takes effect.
// After Shutdown, Init can be called again to reinitialize.
func Init(cfg Config) {
	shutdownMu.Lock()
	wasShutdown := shutdown
	shutdownMu.Unlock()

	if wasShutdown {
		// Reset for re-initialization
		initOnce = sync.Once{}
	}

	initOnce.Do(func() {
		initMu.Lock()
		defer initMu.Unlock()
		agent, err := newAgent(cfg)
		if err != nil {
			panic(fmt.Sprintf("failed to initialize logger: %v", err))
		}
		globalAgent = agent
		shutdownMu.Lock()
		shutdown = false
		shutdownMu.Unlock()
	})
}

// Shutdown gracefully shuts down the logger.
func Shutdown(ctx context.Context) {
	shutdownMu.Lock()
	defer shutdownMu.Unlock()
	if shutdown {
		return
	}
	shutdown = true

	initMu.RLock()
	agent := globalAgent
	initMu.RUnlock()

	if agent != nil {
		agent.stop(ctx)
		initMu.Lock()
		globalAgent = nil
		initMu.Unlock()
	}
}

// IsInitialized returns whether the logger has been initialized.
func IsInitialized() bool {
	initMu.RLock()
	defer initMu.RUnlock()
	return globalAgent != nil
}

// log enqueues a log event at the specified level.
func log(level Level, iface, msg string, params ...interface{}) {
	if iface == "" {
		iface = defaultIface
	}

	initMu.RLock()
	agent := globalAgent
	initMu.RUnlock()

	if agent == nil {
		return
	}

	event := Event{
		Level:   level,
		Iface:   iface,
		Message: msg,
		Params:  params,
	}

	if agent.enqueue(event) {
		recordAccepted()
	} else {
		recordDrop(level)
	}
}

// Debug logs a debug message.
func Debug(iface, msg string, params ...interface{}) {
	log(LevelDebug, iface, msg, params...)
}

// Info logs an info message.
func Info(iface, msg string, params ...interface{}) {
	log(LevelInfo, iface, msg, params...)
}

// Success logs a success message.
func Success(iface, msg string, params ...interface{}) {
	log(LevelSuccess, iface, msg, params...)
}

// Warning logs a warning message.
func Warning(iface, msg string, params ...interface{}) {
	log(LevelWarning, iface, msg, params...)
}

// Fail logs a fail message.
func Fail(iface, msg string, params ...interface{}) {
	log(LevelFail, iface, msg, params...)
}

// Error logs an error message.
func Error(iface, msg string, params ...interface{}) {
	log(LevelError, iface, msg, params...)
}

// Catastrophe logs a catastrophe message.
func Catastrophe(iface, msg string, params ...interface{}) {
	log(LevelCatastrophe, iface, msg, params...)
}

// AudioWritePCM16 writes PCM16 frames to the audio log.
func AudioWritePCM16(frames []int16) {
	initMu.RLock()
	agent := globalAgent
	initMu.RUnlock()

	if agent != nil && agent.audioWriter != nil {
		agent.audioWriter.WritePCM16(frames)
	}
}

// AudioWriteBytesPCM16LE writes PCM16 little-endian bytes to the audio log.
func AudioWriteBytesPCM16LE(data []byte) {
	initMu.RLock()
	agent := globalAgent
	initMu.RUnlock()

	if agent != nil && agent.audioWriter != nil {
		agent.audioWriter.WriteBytesPCM16LE(data)
	}
}


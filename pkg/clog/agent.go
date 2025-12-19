// Package clog: agent goroutine that processes log events.
package clog

import (
	"context"
	"fmt"
	"sync"

	"github.com/LastBotInc/coralie-logging-go/pkg/pcmlog"
)

// agent manages the logging agent goroutine and processes log events.
type agent struct {
	cfg        Config
	queue      chan Event
	done       chan struct{}
	wg         sync.WaitGroup
	shutdown   bool
	mu         sync.Mutex
	consoleSink *consoleSink
	fileSink   *fileSink
	dedupe     *dedupeState
	audioWriter interface {
		WritePCM16([]int16) error
		WriteBytesPCM16LE([]byte) error
		Flush() error
		Close() error
	}
}

// newAgent creates a new agent with the given configuration.
func newAgent(cfg Config) (*agent, error) {
	a := &agent{
		cfg:   cfg,
		queue: make(chan Event, cfg.QueueSize),
		done:  make(chan struct{}),
	}

	// Initialize console sink
	if cfg.Console.Enabled {
		a.consoleSink = newConsoleSink(cfg.Console)
	}

	// Initialize file sink
	if cfg.File.BaseDir != "" {
		var err error
		a.fileSink, err = newFileSink(cfg.File)
		if err != nil {
			return nil, err
		}
	}

	// Initialize deduplication
	a.dedupe = newDedupeState(cfg.Dedupe)

	// Initialize audio writer
	if cfg.Audio.Enabled {
		pcmlogCfg := pcmlog.Config{
			Enabled:         cfg.Audio.Enabled,
			SampleRate:      cfg.Audio.SampleRate,
			Channels:        cfg.Audio.Channels,
			BitsPerSample:   cfg.Audio.BitsPerSample,
			OutputDir:       cfg.Audio.OutputDir,
			FilenamePattern: cfg.Audio.FilenamePattern,
		}
		audioWriter, err := pcmlog.NewWriter(pcmlogCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create audio writer: %w", err)
		}
		if audioWriter != nil {
			a.audioWriter = audioWriter
		}
	}

	a.wg.Add(1)
	go a.run()
	return a, nil
}

// enqueue attempts to enqueue an event, applying drop policy if queue is full.
func (a *agent) enqueue(e Event) bool {
	a.mu.Lock()
	shutdown := a.shutdown
	a.mu.Unlock()
	if shutdown {
		return false
	}

	select {
	case a.queue <- e:
		return true
	default:
		// Queue is full, apply drop policy
		if a.cfg.DropPolicy == "drop_old" {
			// Try to drop oldest by receiving one
			select {
			case <-a.queue:
				// Dropped one, now try to enqueue
				select {
				case a.queue <- e:
					return true
				default:
					return false
				}
			default:
				return false
			}
		}
		// drop_new: just drop the new event
		return false
	}
}

// run is the main agent loop that processes events.
func (a *agent) run() {
	defer a.wg.Done()

	for {
		select {
		case <-a.done:
			// Drain remaining events
			for {
				select {
				case e := <-a.queue:
					a.processEvent(e)
				default:
					// Flush dedupe summary before exiting
					a.flushDedupeSummary()
					return
				}
			}
		case e := <-a.queue:
			a.processEvent(e)
		}
	}
}

// processEvent processes a single log event.
func (a *agent) processEvent(e Event) {
	// Format the message first
	formatted := a.formatMessage(e)

	// Check deduplication
	shouldSuppress, shouldEmitSummary := a.dedupe.check(e.Level, e.Iface, formatted)

	// Emit summary if needed
	if shouldEmitSummary {
		a.emitDedupeSummary()
	}

	// Suppress duplicate message
	if shouldSuppress {
		return
	}

	// Call hooks before processing
	a.callHooks(e)

	// Write to console sink
	if a.consoleSink != nil {
		a.consoleSink.write(e.Level, e.Iface, formatted)
	}

	// Write to file sink
	if a.fileSink != nil {
		a.fileSink.write(e.Level, e.Iface, formatted)
	}

	// Record emitted
	recordEmitted()
}

// emitDedupeSummary emits a deduplication summary message.
func (a *agent) emitDedupeSummary() {
	level, iface, summary, ok := a.dedupe.flushSummary()
	if !ok {
		return
	}

	// Create summary event
	summaryEvent := Event{
		Level:   level,
		Iface:   iface,
		Message: summary,
		Params:  nil,
	}

	// Call hooks
	a.callHooks(summaryEvent)

	// Write to sinks
	if a.consoleSink != nil {
		a.consoleSink.write(level, iface, summary)
	}
	if a.fileSink != nil {
		a.fileSink.write(level, iface, summary)
	}

	recordEmitted()
}

// flushDedupeSummary flushes any pending deduplication summary.
func (a *agent) flushDedupeSummary() {
	a.emitDedupeSummary()
}

// callHooks invokes all applicable hooks for the event.
func (a *agent) callHooks(e Event) {
	// Call global hooks
	if a.cfg.Hooks.Global != nil {
		for _, hook := range a.cfg.Hooks.Global {
			hook.OnLog(e)
		}
	}

	// Call per-level hooks
	if a.cfg.Hooks.PerLevel != nil {
		if hooks, ok := a.cfg.Hooks.PerLevel[e.Level]; ok {
			for _, hook := range hooks {
				hook.OnLog(e)
			}
		}
	}
}

// formatMessage formats a log event message (message part only, no prefix).
func (a *agent) formatMessage(e Event) string {
	if len(e.Params) > 0 {
		return fmt.Sprintf(e.Message, e.Params...)
	}
	return e.Message
}

// stop stops the agent and drains the queue.
func (a *agent) stop(ctx context.Context) {
	a.mu.Lock()
	alreadyShutdown := a.shutdown
	if !alreadyShutdown {
		a.shutdown = true
		close(a.done)
	}
	a.mu.Unlock()

	if alreadyShutdown {
		return
	}

	// Wait for agent to finish or context timeout
	done := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}

	// Flush and close sinks
	if a.fileSink != nil {
		a.fileSink.flush()
		a.fileSink.close()
	}

	// Flush and close audio writer
	if a.audioWriter != nil {
		a.audioWriter.Flush()
		a.audioWriter.Close()
	}
}


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
	cfg         Config
	queue       chan Event
	done        chan struct{}
	wg          sync.WaitGroup
	shutdown    bool
	mu          sync.Mutex
	sinks       []Sink
	dedupe      *dedupeState
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

	// Build sinks: console and file from existing config (backward compatible)
	if cfg.Console.Enabled {
		a.sinks = append(a.sinks, newConsoleSink(cfg.Console))
	}
	if cfg.File.BaseDir != "" {
		fs, err := newFileSink(cfg.File)
		if err != nil {
			return nil, err
		}
		if fs != nil {
			a.sinks = append(a.sinks, fs)
		}
	}
	// Additional sinks from Config.Sinks (e.g. BetterStack) are added in buildExtraSinks
	extra, err := buildExtraSinks(cfg.Sinks)
	if err != nil {
		return nil, err
	}
	a.sinks = append(a.sinks, extra...)

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

// buildExtraSinks builds sinks from Config.Sinks (e.g. BetterStack). Returns nil slice on no config or error.
func buildExtraSinks(cfgs []SinkConfig) ([]Sink, error) {
	if len(cfgs) == 0 {
		return nil, nil
	}
	var out []Sink
	for _, c := range cfgs {
		switch c.Type {
		case "betterstack":
			s, err := newBetterStackSink(c)
			if err != nil {
				return nil, err
			}
			if s != nil {
				out = append(out, s)
			}
		default:
			// Unknown type: skip (or could return error)
			continue
		}
	}
	return out, nil
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
//
// Ordering is security-critical (LAS-1488, Gemini + CodeRabbit review):
//
//  1. Bound oversized string params BEFORE formatting. A multi-megabyte %s arg
//     would otherwise force a giant fmt.Sprintf allocation/concat on the agent
//     goroutine before any truncation. boundParams truncates such params (with a
//     marker) up front, capping per-call work; Redact's own length guard stays as
//     a backstop. The caller's Params slice is never mutated.
//  2. Format the (bounded) message.
//  3. Dedupe on the RAW formatted string. Two distinct callers that differ only
//     in PII (participant_id=1111111@.. vs 2222222@..) must NOT collapse into one
//     dedupe key, or the second caller is silently suppressed -- blinding ops to
//     concurrent calls. So the dedupe key is computed BEFORE redaction.
//  4. Only after the suppress check passes do we redact the formatted string once.
//     Hooks receive an Event whose Message is that single redacted string with
//     Params cleared -- so a hook that re-formats or serializes the Event cannot
//     recombine PII split across Message+Params (e.g. "%s@%s" + ["alice","x.com"]).
//     The same redacted string feeds every sink.
func (a *agent) processEvent(e Event) {
	// Bound oversized string params before formatting (DoS guard, pre-Sprintf).
	e.Params = boundParams(e.Params)

	// Format the (bounded) message.
	formatted := a.formatMessage(e)

	// Check deduplication on the RAW (pre-redaction) formatted string so distinct
	// callers are not collapsed by redaction tokens.
	shouldSuppress, shouldEmitSummary := a.dedupe.check(e.Level, e.Iface, formatted)

	// Emit summary if needed.
	if shouldEmitSummary {
		a.emitDedupeSummary()
	}

	// Suppress duplicate message.
	if shouldSuppress {
		return
	}

	// Centralized PII redaction (LAS-1488 layer #1). Redact the formatted string
	// once, only after the dedupe suppress check, so neither the sinks nor the
	// hooks ever see raw caller PII.
	formattedRedacted := Redact(formatted)

	// Hand hooks the redacted formatted string with Params cleared, so a hook that
	// re-formats/serializes the Event cannot reconstruct PII from the fragments.
	hookEvent := e
	hookEvent.Message = formattedRedacted
	hookEvent.Params = nil
	a.callHooks(hookEvent)

	for _, sink := range a.sinks {
		sink.Write(e.Level, e.Iface, formattedRedacted)
	}

	// Record emitted
	recordEmitted()
}

// boundParams returns a Params slice in which every string longer than
// maxRedactLen has been truncated (with a "…<truncated N bytes>" marker) so the
// downstream fmt.Sprintf in formatMessage cannot be forced into a multi-megabyte
// allocation/concat by an attacker-controlled %s arg. The caller's slice is never
// mutated: a fresh slice is allocated only when at least one param needs bounding;
// otherwise the original slice is returned unchanged (no allocation, common case).
func boundParams(params []interface{}) []interface{} {
	for i, p := range params {
		s, ok := p.(string)
		if !ok || len(s) <= maxRedactLen {
			continue
		}
		// At least one oversized string: copy, then bound from here on.
		out := make([]interface{}, len(params))
		copy(out, params)
		for j := i; j < len(out); j++ {
			if s, ok := out[j].(string); ok && len(s) > maxRedactLen {
				out[j] = boundString(s)
			}
		}
		return out
	}
	return params
}

// emitDedupeSummary emits a deduplication summary message.
//
// The default summaryFormat ("last message repeated %d more times") is count-only
// and never interpolates the stored raw message, so the raw formatted string held
// in dedupe state is never written out. summaryFormat is configurable, though, so
// we route the summary through the same redaction the normal path uses -- both the
// sink string and the hook Event -- as a defensive guarantee that no future format
// (or a misconfigured %s) can leak the in-memory raw message to a sink or hook.
func (a *agent) emitDedupeSummary() {
	level, iface, summary, ok := a.dedupe.flushSummary()
	if !ok {
		return
	}

	summaryRedacted := Redact(summary)

	// Hand hooks the redacted summary string (Params already nil for summaries),
	// matching processEvent: hooks never see a raw, reconstructable message.
	summaryEvent := Event{
		Level:   level,
		Iface:   iface,
		Message: summaryRedacted,
		Params:  nil,
	}

	// Call hooks
	a.callHooks(summaryEvent)

	for _, sink := range a.sinks {
		sink.Write(level, iface, summaryRedacted)
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

	for _, sink := range a.sinks {
		sink.Flush()
		sink.Close()
	}

	// Flush and close audio writer
	if a.audioWriter != nil {
		_ = a.audioWriter.Flush()
		_ = a.audioWriter.Close()
	}
}

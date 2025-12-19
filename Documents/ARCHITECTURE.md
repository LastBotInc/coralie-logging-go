# Architecture

## Overview

coralie-logging-go is an asynchronous, high-performance logging library built around a single agent goroutine that processes log events from a bounded queue.

## Core Components

### Event Flow

1. **API Calls** (`pkg/clog/api.go`): User code calls `clog.Info()`, `clog.Error()`, etc.
2. **Event Creation**: Events are created with level, interface, message, and parameters
3. **Queue**: Events are enqueued to a bounded channel (configurable size)
4. **Agent Goroutine**: Single goroutine processes events sequentially
5. **Formatting**: Messages are formatted in the agent goroutine (using `fmt.Appendf`)
6. **Sinks**: Formatted messages are written to configured sinks (console, files)
7. **Hooks**: Hooks are invoked in the agent goroutine before writing

### Key Design Decisions

- **Single Writer**: All formatting and writing happens in one goroutine to avoid races
- **Bounded Queue**: Prevents unbounded memory growth; configurable drop policy
- **Formatting in Agent**: Reduces per-call allocations; formatting happens asynchronously
- **Deduplication**: Collapses consecutive identical messages to reduce noise

## Package Structure

### `pkg/clog`

Main logging package providing:
- Public API functions (Debug, Info, Success, etc.)
- Configuration and initialization
- Event model and levels
- Agent goroutine and queue management
- Sinks (console, file)
- Hooks system
- Deduplication logic
- Shutdown and signal handling
- Statistics tracking

### `pkg/pcmlog`

Audio PCM/WAV logging package:
- PCM16 frame writing
- WAV file generation
- Buffered writes with flush on shutdown

### `internal/term`

Terminal utilities:
- TTY detection
- Color support detection
- Color formatting

### `internal/timefmt`

Time formatting utilities for log timestamps.

## Thread Safety

- All public API functions are safe for concurrent use
- Internal state is protected by the single agent goroutine
- Statistics use atomic operations where needed

## Performance Characteristics

- Minimal per-call allocations (only event struct and queue send)
- Formatting deferred to agent goroutine
- Bounded queue prevents memory leaks
- Drop policy prevents blocking on slow sinks

See [PERFORMANCE.md](PERFORMANCE.md) for detailed performance notes.


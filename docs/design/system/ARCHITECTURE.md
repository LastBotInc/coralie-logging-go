# System Architecture

## Overview

`coralie-logging-go` is a shared Go library providing structured, high-performance logging with deduplication, audio PCM/WAV support, and graceful shutdown.

## Component Diagram

```
Caller goroutine(s)           Agent goroutine
  |                              |
  | clog.Info(iface, msg)        |
  |---> [bounded queue] -------->|
  |                              |---> dedupe filter
  |                              |---> format (fmt.Appendf)
  |                              |---> console sink
  |                              |---> file sink(s)
  |                              |---> betterstack sink
  |                              |---> hooks (global + per-level)
  |                              |---> stats update
  |                              |
  | clog.Shutdown(ctx)           |
  |---> drain queue ------------>|---> flush all sinks
                                 |---> close file handles
                                 |---> finalize WAV files
                                 |---> stop goroutine
```

## Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| Single agent goroutine | Eliminates races; all formatting and I/O serialized |
| Bounded queue + drop policy | Prevents OOM under log storms |
| Formatting in agent | Reduces caller-side allocations |
| sync.Pool for buffers | Reuse event buffers, minimize GC pressure |
| Interface-based sinks | Extensible without modifying core |

## Packages

| Package | Responsibility |
|---------|---------------|
| `pkg/clog` | Core API, config, agent, sinks, dedupe, hooks, shutdown |
| `pkg/pcmlog` | PCM frame to WAV file writer |
| `internal/term` | TTY detection, ANSI color codes |
| `internal/timefmt` | Timestamp formatting |

## Data Flow

1. Caller calls `clog.Info("App", "msg %s", arg)`
2. Event created (level, iface, raw msg, params, timestamp)
3. Event pushed to bounded channel (dropped if full per policy)
4. Agent goroutine dequeues event
5. Dedupe filter: suppress if identical to previous, count repeats
6. Format message using `fmt.Appendf` into pooled buffer
7. Write to all enabled sinks (console, file, betterstack)
8. Invoke hooks
9. Update stats (accepted, emitted, drops)

## Shutdown Sequence

1. `Shutdown(ctx)` called
2. Close event channel (no new events accepted)
3. Agent drains remaining events from channel
4. Flush dedupe summary if pending
5. Flush all sink buffers
6. Close file handles
7. Finalize WAV files (write headers)
8. Agent goroutine exits
9. `Shutdown` returns

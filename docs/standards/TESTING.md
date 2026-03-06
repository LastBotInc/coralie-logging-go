# Testing Standards

## Coverage Requirements

- All exported functions must have unit tests
- All behavior paths (happy path, error, edge cases) must be tested
- No test may require real hardware (microphone, network services)

## Test Commands

```bash
# From repo root - all library tests
go test ./...

# With race detector
go test -race ./...

# Verbose output
go test -v ./...

# Specific package
go test ./pkg/clog/...

# Example module (separate go.mod)
cd examples/fyne-audio-monitor && go test ./...
cd examples/fyne-audio-monitor && go test -race ./...
```

## Test Organization

| Pattern | Convention |
|---------|------------|
| File naming | `*_test.go` colocated with source |
| Test naming | `TestFunctionName_Scenario` |
| Table-driven | Preferred for multiple input/output cases |
| Interfaces | Use for external dependencies (audio, network) |
| Fakes | Provide fake implementations for hardware dependencies |

## Required Test Categories

| Category | What to Test |
|----------|-------------|
| Dedupe | Consecutive collapse, non-consecutive pass-through, flush on shutdown |
| Drop policy | Small queue, spam, assert drops counted |
| File routing | Temp dir, verify correct files written |
| Hooks | Global and per-level invocation |
| Shutdown | Drain queue, flush sinks, no goroutine leaks |
| Audio writer | Creates non-empty WAV file (no mic required) |
| Stats | Drop counts, accepted/emitted counts |

## Test Isolation

- Use `t.TempDir()` for file-based tests
- Use interfaces + fakes for hardware dependencies
- Provide `AudioSource` interface with `FakeSource` for deterministic testing
- Tests must be parallelizable where possible

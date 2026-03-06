# Go Testing Reference

## Commands

```bash
# All tests
go test ./...

# With race detector (required before merge)
go test -race ./...

# Verbose
go test -v ./...

# Specific package
go test -v ./pkg/clog/...
go test -v ./pkg/pcmlog/...
go test -v ./internal/term/...
go test -v ./internal/timefmt/...

# Specific test
go test -v -run TestDedupe_CollapsesConsecutive ./pkg/clog/

# Coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Benchmarks
go test -bench=. ./pkg/clog/
```

## Example Module Tests

```bash
cd examples/fyne-audio-monitor
go test ./...
go test -race ./...
```

## Test File Map

| Test File | Tests |
|-----------|-------|
| `pkg/clog/agent_test.go` | Agent lifecycle, drop policy |
| `pkg/clog/api_test.go` | Public API functions |
| `pkg/clog/betterstack_sink_test.go` | BetterStack sink |
| `pkg/clog/config_test.go` | Config defaults and validation |
| `pkg/clog/dedupe_test.go` | Deduplication behavior |
| `pkg/clog/format_test.go` | Log formatting |
| `pkg/clog/hooks_test.go` | Hook invocation |
| `pkg/clog/levels_test.go` | Level string/ordering |
| `pkg/clog/shutdown_test.go` | Shutdown, drain, flush |
| `pkg/clog/sinks_test.go` | Sink integration |
| `pkg/pcmlog/writer_test.go` | WAV file creation |
| `internal/term/color_test.go` | Color code output |
| `internal/term/tty_test.go` | TTY detection |
| `internal/timefmt/timefmt_test.go` | Time formatting |

## Test Patterns

- Table-driven tests for multiple input/output combinations
- `t.TempDir()` for file system tests
- Interfaces + fakes for hardware (audio capture)
- `context.WithTimeout` for shutdown tests
- `sync.WaitGroup` for goroutine leak detection

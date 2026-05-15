# Go Policies

## Language & Toolchain

- Go version: 1.26.3+ (as specified in go.mod)
- Module path: `github.com/LastBotInc/coralie-logging-go`
- No vendoring; use Go module proxy

## Code Style

- `gofmt` formatted (enforced)
- Every exported identifier has GoDoc comment
- Every file starts with header comment: `// Package foo: brief description.`
- Prefer `fmt.Appendf` over `fmt.Sprintf` in hot paths
- Use `sync.Pool` for reusable buffers (event buffers, param slices)

## Architecture Patterns

| Pattern | Convention |
|---------|------------|
| Public API | Top-level functions in `pkg/clog/api.go` |
| Sinks | Implement `Sink` interface in dedicated `*_sink.go` files |
| Config | Struct fields in `config.go`, defaults in `DefaultConfig()` |
| Agent loop | Single writer goroutine processes all events (no races) |
| Formatting | Done in agent goroutine, not in caller goroutine |
| Shutdown | `Shutdown(ctx)` drains queue, flushes sinks, closes handles |

## Package Layout

```
pkg/clog/       -- core logging API (public)
pkg/pcmlog/     -- PCM/WAV audio logging (public)
internal/term/  -- TTY detection, color codes (private)
internal/timefmt/ -- time formatting (private)
cmd/            -- demo CLI programs
examples/       -- example applications (separate go.mod)
```

## Build & Test Commands

```bash
# Build
go build ./...

# Test (all)
go test ./...

# Test with race detector
go test -race ./...

# Test specific package
go test -v ./pkg/clog/...

# Run demo
go run ./cmd/coralie-logging-demo

# Version management
make version
make bump-patch
make bump-minor
make bump-major

# Dependencies
go mod tidy
go mod verify
```

## Performance Rules

- No per-log-call heap churn in steady state
- Bounded queue with configurable size and drop policy
- Track drops in stats (per-level)
- Single writer goroutine handles all sinks
- Use `sync.Pool` for event buffers

## Error Handling

- Return errors from initialization functions
- Log-and-continue for sink write failures
- Never panic in library code (except catastrophe-level if configured)
- Provide `RecoverAndFlush(repanic bool)` for caller panic handling

## Testing Rules

- All exported functions must have tests
- No test may require real hardware
- Use interfaces + fakes for external dependencies
- Use `t.TempDir()` for file tests
- Table-driven tests preferred for multiple cases

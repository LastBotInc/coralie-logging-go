# Go Development Workflow

## Setup

```bash
# Clone
git clone https://github.com/LastBotInc/coralie-logging-go.git
cd coralie-logging-go

# Verify Go version
go version  # must be 1.24+

# Download dependencies
go mod download

# Verify everything compiles
go build ./...

# Run tests
go test ./...
```

## Development Cycle

1. Create feature branch: `git checkout -b feature/my-change`
2. Write code following [GOLANG_POLICIES.md](../../policies/GOLANG_POLICIES.md)
3. Write tests colocated with source (`*_test.go`)
4. Run tests: `go test -race ./...`
5. Update documentation (mandatory)
6. Commit with clear message
7. Create PR

## Package Structure

| Package | Visibility | Purpose |
|---------|-----------|---------|
| `pkg/clog` | Public | Core logging API |
| `pkg/pcmlog` | Public | PCM/WAV audio logging |
| `internal/term` | Private | TTY detection, ANSI colors |
| `internal/timefmt` | Private | Time formatting utilities |

## Adding a New Sink

1. Create `pkg/clog/newsink_sink.go`
2. Implement the `Sink` interface
3. Add config fields to `SinkConfig`
4. Register in agent loop
5. Write tests in `pkg/clog/newsink_sink_test.go`
6. Update `Documents/CONFIGURATION.md` and `Documents/ARCHITECTURE.md`

## Adding a New Log Level

1. Add constant in `pkg/clog/levels.go`
2. Add `String()` case
3. Add public API function in `pkg/clog/api.go`
4. Add color mapping in `internal/term/color.go`
5. Update `Documents/LEVELS.md`

## Debugging

```bash
# Verbose test output
go test -v -run TestSpecificTest ./pkg/clog/

# Race detector
go test -race ./pkg/clog/

# CPU/memory profiling
go test -bench=. -cpuprofile=cpu.prof ./pkg/clog/
go tool pprof cpu.prof
```

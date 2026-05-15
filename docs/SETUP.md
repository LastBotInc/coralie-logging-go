# Setup Guide

## Prerequisites

| Tool | Version | Check |
|------|---------|-------|
| Go | 1.26.3+ | `go version` |
| Git | any | `git --version` |
| Make | any | `make --version` |

## Quick Start

```bash
# Clone the repository
git clone https://github.com/LastBotInc/coralie-logging-go.git
cd coralie-logging-go

# Download dependencies
go mod download

# Verify build
go build ./...

# Run tests
go test ./...

# Run demo
go run ./cmd/coralie-logging-demo
```

## Project Structure

```
coralie-logging-go/
  go.mod                          # Module definition
  Makefile                        # Version management targets
  pkg/clog/                       # Core logging API (public)
  pkg/pcmlog/                     # Audio PCM/WAV logging (public)
  internal/term/                  # TTY detection, colors (private)
  internal/timefmt/               # Time formatting (private)
  cmd/coralie-logging-demo/       # Demo CLI
  examples/fyne-audio-monitor/    # Example app (separate module)
  Documents/                      # Project documentation
  docs/                           # Claude agent documentation
  task/                           # Task tracking
  scripts/                        # Automation scripts
```

## IDE Setup

No special IDE configuration required. Standard Go tooling works:

```bash
# Format code
gofmt -w .

# Vet code
go vet ./...
```

## Example Module

The `examples/fyne-audio-monitor/` has its own `go.mod` with a `replace` directive for local development:

```bash
cd examples/fyne-audio-monitor
go mod download
go test ./...
```

## Generating CLAUDE.md

```bash
bash scripts/setup-stacks.sh golang
```

This generates `CLAUDE.md` at the project root from policies and stack configuration.

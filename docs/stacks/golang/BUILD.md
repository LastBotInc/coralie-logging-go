# Go Build Reference

## Prerequisites

- Go 1.24 or later
- No CGO dependencies for core library

## Build Commands

```bash
# Compile all packages
go build ./...

# Compile specific package
go build ./pkg/clog/...

# Build demo binary
go build -o bin/demo ./cmd/coralie-logging-demo

# Cross-compile
GOOS=linux GOARCH=amd64 go build ./...
```

## Module Management

```bash
# Initialize (already done)
# go mod init github.com/LastBotInc/coralie-logging-go

# Add dependency
go get github.com/example/pkg@v1.0.0

# Remove unused / add missing
go mod tidy

# Verify checksums
go mod verify

# View dependency graph
go mod graph
```

## Release Process

| Step | Command |
|------|---------|
| Sync with remote | `make update` |
| Check current version | `make version` |
| Patch release | `make bump-patch` |
| Minor release | `make bump-minor` |
| Major release | `make bump-major` |

All `bump-*` targets create an annotated git tag and push to origin.

## CI/CD

- GitHub Actions triggers on push to `main` (staging) and version tags (production)
- Docker images pushed to `ghcr.io/lastbotinc/coralie-logging-go`
- ArgoCD picks up production tags via semver strategy

## Build Constraints

- No build tags required for core library
- Example `fyne-audio-monitor` requires CGO for Fyne UI framework
- Platform-specific audio capture isolated behind build tags where needed

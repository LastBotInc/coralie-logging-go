# Go Stack Reference

## Overview

| Property | Value |
|----------|-------|
| Language | Go 1.26.3+ |
| Module | `github.com/LastBotInc/coralie-logging-go` |
| Type | Shared library (logging package) |
| Packages | `pkg/clog`, `pkg/pcmlog` |
| Dependencies | `golang.org/x/sys` (indirect) |

## Quick Reference

| Task | Command |
|------|---------|
| Build | `go build ./...` |
| Test | `go test ./...` |
| Test + race | `go test -race ./...` |
| Run demo | `go run ./cmd/coralie-logging-demo` |
| Version | `make version` |
| Release | `make bump-patch` / `bump-minor` / `bump-major` |

## Stack Documents

| Document | Content |
|----------|---------|
| [BUILD.md](BUILD.md) | Build configuration and commands |
| [DEVELOPMENT.md](DEVELOPMENT.md) | Development workflow |
| [TESTING.md](TESTING.md) | Testing practices |
| [SECURITY.md](SECURITY.md) | Security practices |

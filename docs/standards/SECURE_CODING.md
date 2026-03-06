# Secure Coding Standards

## General Principles

- No secrets in source code or logs
- Validate all external inputs before processing
- Use bounded resources (queues, buffers) to prevent resource exhaustion
- Handle errors explicitly; never silently swallow errors

## Go-Specific Security

| Area | Requirement |
|------|-------------|
| Goroutines | Always ensure goroutines can be stopped (context, done channel) |
| File I/O | Use `os.OpenFile` with explicit permissions (0644 files, 0755 dirs) |
| Race conditions | All shared state protected by mutex or channel; pass `go test -race` |
| Input validation | Validate config values (QueueSize > 0, valid DropPolicy values) |
| Panic recovery | `clog.RecoverAndFlush()` available for deferred panic handling |

## Logging Security

- Never log sensitive data (tokens, passwords, PII)
- BetterStack sink tokens must come from environment, not hardcoded
- File paths must be sanitized before use in file sink
- Log rotation/cleanup is caller's responsibility; library does not delete files

## Dependency Security

- Minimal dependencies (currently only `golang.org/x/sys`)
- Audit new dependencies before adding
- Pin dependency versions in go.mod
- Run `go mod verify` to check integrity

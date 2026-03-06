# Go Security Practices

## Race Safety

- All shared state protected by mutex or channel
- Single-writer agent goroutine pattern eliminates most races
- Must pass: `go test -race ./...`

## Resource Bounds

| Resource | Protection |
|----------|-----------|
| Log queue | Bounded by `QueueSize` config (default 1000) |
| Drop policy | `drop_new` (default) or `drop_old` prevents unbounded growth |
| File handles | Closed on `Shutdown()` |
| Goroutines | Stopped on `Shutdown()` via context/done channel |
| Audio buffers | Flushed and finalized on shutdown |

## Input Validation

- Config values validated at `Init()` time
- Invalid `DropPolicy` values rejected
- `QueueSize` must be positive
- File paths used as-is (caller responsibility to sanitize)

## Secrets

- Never log tokens, passwords, or PII through clog
- BetterStack sink token: pass via config from environment variable
- No secrets in source code, tests, or documentation

## Vulnerability Scanning

```bash
# Install govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest

# Scan
govulncheck ./...
```

## Signal Handling

- `clog.InstallSignalHandler()` catches SIGINT/SIGTERM
- Triggers graceful shutdown: drain queue, flush sinks, close files
- `clog.RecoverAndFlush(repanic)` for panic recovery in deferred calls

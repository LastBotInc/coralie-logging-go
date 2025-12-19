# Shutdown, Panic Recovery, and Signal Handling

## Graceful Shutdown

### Manual Shutdown

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
clog.Shutdown(ctx)
```

`Shutdown()`:
1. Stops accepting new events
2. Drains the queue
3. Flushes all sinks (console, files, audio)
4. Closes file handles
5. Finalizes WAV files
6. Stops all goroutines

### Signal Handling

Install automatic signal handler:

```go
stop := clog.InstallSignalHandler()
defer stop()
```

Handles `SIGINT` and `SIGTERM`, calls `Shutdown()` with a default timeout.

## Panic Recovery

Recover from panics and flush logs:

```go
defer clog.RecoverAndFlush(true) // true = re-panic after flush
```

`RecoverAndFlush()`:
1. Recovers from panic
2. Logs the panic message
3. Flushes all sinks
4. Optionally re-panics

## Best Practices

1. Always call `Shutdown()` before program exit
2. Use `defer clog.Shutdown(ctx)` in main
3. Install signal handler for long-running services
4. Use `RecoverAndFlush()` in goroutines that might panic

## Timeout

Shutdown uses context timeout. If timeout expires:
- Remaining events may be dropped
- File handles may not be closed gracefully
- Set appropriate timeout based on queue size and sink performance


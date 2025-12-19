# Performance

## Design Goals

- Minimal per-call allocations
- Non-blocking API calls
- Bounded memory usage
- Efficient formatting

## Allocation Strategy

### Per-Call Allocations

Each log call allocates:
- Event struct (small, stack-allocated if possible)
- Queue send operation (minimal overhead)

Formatting happens asynchronously in the agent goroutine, not during the API call.

### Formatting

- Uses `fmt.Appendf` for efficient string building
- Formatting happens in agent goroutine (serial, no races)
- Reuses buffers where possible

### Queue Management

- Bounded channel prevents unbounded growth
- Drop policy prevents blocking
- Configurable size based on workload

## Performance Characteristics

### Throughput

- High throughput for console-only logging
- File I/O is the main bottleneck
- Audio logging adds minimal overhead when disabled

### Latency

- API calls return immediately (non-blocking)
- Actual write latency depends on queue depth and sink performance

### Memory

- Bounded by queue size
- Formatting buffers are reused
- No per-message heap allocations in steady state

## Optimization Tips

1. **Queue Size**: Set based on burst capacity needed
2. **Drop Policy**: Use `drop_old` for high-throughput scenarios
3. **File Sinks**: Disable file logging if not needed
4. **Deduplication**: Reduces I/O for repetitive messages
5. **Level Filtering**: Omit unnecessary levels from expensive sinks

## Benchmarks

Run benchmarks with:
```bash
go test -bench=. -benchmem ./pkg/clog
```


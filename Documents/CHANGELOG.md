# Changelog

All notable changes to coralie-logging-go will be documented in this file.

## [Unreleased]

### Added
- Centralized PII redaction (LAS-1488 layer #1): every formatted log message is
  scrubbed once in the agent's `processEvent` choke point before it reaches
  dedupe and all sinks (console, file, BetterStack/Postgres). New `pkg/clog/redact.go`
  with an ordered, precompiled `Redactor`: email -> `<email>`, IPv4 (optional
  `:port`, incl. media IPs) -> `<ip>`, `+E.164` phone -> `<phone>`, and the
  `<CID>@<ip>` / `<DID>@<ip>` participant/conference-id phone form -> `<phone>@`.
  Bare in-sentence digit runs are intentionally left untouched to avoid clobbering
  timestamps, ports, sample/byte/frame counts, UUIDs, and version strings.
  Enabled by default; disable for local dev via `CORALIE_LOG_REDACT=0` (also
  accepts `false`/`no`/`off`) or `SetRedactionEnabled(false)`. Patterns are
  replaceable via `NewRedactor` + `SetRedactor` for ops tuning. Added
  `redact_test.go` (true-positive + false-positive guards, idempotence, toggle,
  full-sink-path wiring test) and `BenchmarkRedact`.
- Initial repository structure
- Documentation stubs
- Package skeletons
- Core logging API with async agent goroutine
- Bounded queue with configurable drop policies (drop_new, drop_old)
- Statistics tracking (drops per level, accepted count, emitted count)
- Init/Shutdown with sync.Once guard and re-initialization support
- All log level functions (Debug, Info, Success, Warning, Fail, Error, Catastrophe)
- Message formatting in agent goroutine using fmt.Appendf
- Comprehensive unit tests including race detection tests
- Console sink with color and emoji support (TTY-aware)
- File sink with per-level routing
- Hooks system (global and per-level)
- Deduplication of consecutive identical messages
- Signal handling (SIGINT, SIGTERM) with graceful shutdown
- Panic recovery with log flushing
- PCM16 audio logging to WAV files
- Demo CLI application demonstrating all features
- Fyne audio monitor example application with separate module


# Changelog

All notable changes to coralie-logging-go will be documented in this file.

## [Unreleased]

### Security
- PII-redaction hardening (LAS-1488, Gemini review). Three fixes in
  `pkg/clog/agent.go` `processEvent` + `pkg/clog/redact.go`:
  - **Hook bypass:** hooks previously received the raw, unredacted `Event`
    (`Message` + `Params` carry caller PII), so any hook forwarding to
    Sentry/Datadog/webhooks leaked PII. New `Redactor.RedactEvent` / package-level
    `RedactEvent` produce a redacted CLONE (scrub `Message` template + every
    string-valued param; non-string params untouched; caller slice never mutated),
    and `callHooks` now receives that clone.
  - **Dedupe collision:** redaction ran before `dedupe.check`, so two distinct
    callers on the same line collapsed to one redacted key and the second was
    suppressed as a "duplicate". `processEvent` now dedupes on the RAW formatted
    string and redacts only after the suppress check passes. Dedupe summary path
    confirmed count-only and additionally routed through redaction defensively.
  - **Payload-size DoS bound:** `Redactor.Redact` now truncates inputs over
    `maxRedactLen` (64 KiB) with a `…<truncated N bytes>` marker before running the
    regexes, capping per-call work from attacker-controlled fields (huge SIP
    headers / malformed packets). PII near the start is still redacted.
  - Added tests: hook-redaction, dedupe-preserves-distinct-callers,
    DoS-bound truncation, `RedactEvent` clone semantics.

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


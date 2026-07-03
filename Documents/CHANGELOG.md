# Changelog

All notable changes to coralie-logging-go will be documented in this file.

## [Unreleased]

## v0.2.0 (2026-06-08)

### Security
- PII-redaction hardening (LAS-1488, Gemini + CodeRabbit review). Fixes in
  `pkg/clog/agent.go` `processEvent` + `pkg/clog/redact.go`:
  - **Hook bypass / reconstruction:** hooks previously received the raw,
    unredacted `Event` (`Message` + `Params` carry caller PII), so any hook
    forwarding to Sentry/Datadog/webhooks leaked PII. Scrubbing the template and
    params independently still let a hook recombine PII split across them (e.g.
    `Message="%s@%s"`, `Params=["alice","example.com"]`). `callHooks` now receives
    an `Event` whose `Message` is the single REDACTED FORMATTED string with
    `Params` cleared (`Level`/`Iface` preserved), so no fragment recombination is
    possible. The reconstruction-prone `Redactor.RedactEvent` / package-level
    `RedactEvent` (unreleased — never shipped in a tagged version) were removed.
  - **Dedupe collision:** redaction ran before `dedupe.check`, so two distinct
    callers on the same line collapsed to one redacted key and the second was
    suppressed as a "duplicate". `processEvent` now dedupes on the RAW formatted
    string and redacts only after the suppress check passes. Dedupe summary path
    confirmed count-only and additionally routed through redaction defensively.
  - **Payload-size DoS bound:** oversized string params are now truncated to
    `maxRedactLen` (64 KiB) with a `…<truncated N bytes>` marker BEFORE
    `fmt.Sprintf` formats them (`boundParams` in `processEvent`), so a
    multi-megabyte `%s` arg can no longer force a giant allocation/concat on the
    agent goroutine. `Redactor.Redact` keeps the same truncation as a backstop.
    PII near the start is still redacted.
  - Split `pkg/clog/redact.go` into `redact.go` (Redactor type + Redact engine +
    truncation) and `redact_config.go` (default patterns, env toggle, package
    globals) to stay under the 200-line file cap; public API unchanged minus the
    removed `RedactEvent`.
  - Added tests: hook receives single redacted formatted Message, hook PII-by-
    formatting reconstruction guard, dedupe-preserves-distinct-callers,
    Redact DoS-bound truncation, agent/sink-path DoS bound (oversized `%s`).

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


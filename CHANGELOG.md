# Changelog

## [Unreleased]

- Add optional `LogContext` trace/span correlation to hooks, Better Stack JSON,
  console and file sinks. Preserve contextless output and existing redaction;
  logging does not initialize an SDK or create spans (LAS-3500).
- 2026-09-15 (LAS-3488): synchronize 55 governance corpus cases and 42 capture
  contract vectors with Rails. Remove positive national-identifier examples
  without verified nonassignment, retain malformed near misses, and use reserved
  email domains. Positive national-ID detection remains unverified by this corpus;
  native redactors and runtime capture behavior are unchanged.

## v0.2.0 (2026-06-08)

Centralized PII redaction layer (LAS-1488 layer #1).

- `pkg/clog` redaction middleware: scrub each formatted log message once in the
  agent `processEvent` choke point, after the dedupe/suppress check and ahead of
  all sinks. Ordered precompiled
  `Redactor`: email → `<email>`, IPv4 (optional `:port`, incl. media IPs) → `<ip>`,
  `+E.164` → `<phone>`, and the `<CID>@<ip>` / `<DID>@<ip>` participant/conference-id
  phone form → `<phone>@`. Default-on; `CORALIE_LOG_REDACT=0/false/no/off` disables.
- Hardening (Gemini + CodeRabbit review): hooks receive a single REDACTED formatted
  message (no cross-field PII recombination); dedupe runs on the raw string before
  redaction (no distinct-caller collision); 64 KiB DoS bound before formatting.
- New exported `Redact`, `NewDefaultRedactor`, `NewRedactor`, `SetRedactor`,
  `SetRedactionEnabled`, `RedactionEnabled`. See Documents/CHANGELOG.md for detail.

## v0.1.0 (2026-03-18)

Initial SemVer release. No API changes from prior untagged state.

- Established SemVer tagging workflow
- Added Makefile with version management targets
- Added gorelease CI for API compatibility checking

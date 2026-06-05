# Configuration

## Overview

Configuration is done via `clog.Config` struct passed to `clog.Init()`.

## Basic Configuration

```go
cfg := clog.DefaultConfig()
cfg.Console.Enabled = true
cfg.File.BaseDir = "./logs"
clog.Init(cfg)
```

## Configuration Options

### Queue

- `QueueSize`: Maximum number of events in queue (default: 1000)
- `DropPolicy`: Behavior when queue is full
  - `"drop_new"`: Drop new events (default)
  - `"drop_old"`: Drop oldest events

### Console Sink

- `Console.Enabled`: Enable console output (default: true)
- `Console.Colors`: Enable colors/emojis (default: auto-detect TTY)
- `Console.OmitLevels`: Map of levels to omit from console

### File Sink

- `File.BaseDir`: Base directory for log files (default: empty, disabled)
- `File.PerLevel`: Map of level to filename (empty string omits level)
  - Example: `map[clog.Level]string{clog.LevelError: "error.log"}`

### Deduplication

- `Dedupe.Enabled`: Enable deduplication (default: true)
- `Dedupe.SummaryFormat`: Format string for summary (default: "last message repeated %d more times")

### Audio Logging

- `Audio.Enabled`: Enable audio PCM/WAV logging (default: false)
- `Audio.SampleRate`: Sample rate in Hz (default: 44100)
- `Audio.Channels`: Number of channels (default: 1)
- `Audio.BitsPerSample`: Bits per sample (default: 16)
- `Audio.OutputDir`: Directory for WAV files (default: "./audio_logs")
- `Audio.FilenamePattern`: Filename pattern with time formatting

### Hooks

- `Hooks.Global`: List of global hooks (called for all levels)
- `Hooks.PerLevel`: Map of level to hooks (called in addition to global)

### Additional Sinks (third-party)

- `Sinks`: Slice of `SinkConfig` for extra sinks (e.g. BetterStack). Nil or empty = no extra sinks.
- Each `SinkConfig` has:
  - `Type`: `"betterstack"` (first supported type)
  - `MinLevel`: Only emit events at or above this level (e.g. `clog.LevelWarning`). Zero (`LevelDebug`) = all levels.
  - `OmitLevels`: Map of levels to omit (same semantics as console)
  - `Format`: `"text"` or `"json"` (BetterStack uses JSON)
  - `Token`: For BetterStack, the source token (required)
  - `Endpoint`: For BetterStack, ingest URL (default `https://in.logs.betterstack.com`)

## Example: Full Configuration

```go
cfg := clog.Config{
    QueueSize: 2000,
    DropPolicy: "drop_old",
    Console: clog.ConsoleConfig{
        Enabled: true,
        Colors: true,
    },
    File: clog.FileConfig{
        BaseDir: "./logs",
        PerLevel: map[clog.Level]string{
            clog.LevelError: "errors.log",
            clog.LevelCatastrophe: "critical.log",
        },
    },
    Dedupe: clog.DedupeConfig{
        Enabled: true,
        SummaryFormat: "[repeated %d times]",
    },
    Audio: clog.AudioConfig{
        Enabled: true,
        SampleRate: 48000,
        Channels: 2,
        OutputDir: "./audio",
    },
}
clog.Init(cfg)
```

## Example: With BetterStack sink

```go
cfg := clog.DefaultConfig()
cfg.Console.Enabled = true
cfg.Sinks = []clog.SinkConfig{
    {
        Type:     "betterstack",
        Token:    os.Getenv("BETTERSTACK_SOURCE_TOKEN"),
        Endpoint: "https://in.logs.betterstack.com", // optional
        MinLevel: clog.LevelWarning,                // only WARNING and above
        Format:   "json",
    },
}
clog.Init(cfg)
```

## Example: Multiple sinks, different levels

```go
cfg.Sinks = []clog.SinkConfig{
    {Type: "betterstack", Token: tok1, MinLevel: clog.LevelInfo, Format: "json"},
    {Type: "betterstack", Token: tok2, MinLevel: clog.LevelError, Format: "json"},
}
```

## PII redaction (LAS-1488)

Every formatted log message is scrubbed for caller PII at a single choke point
inside the agent (`processEvent`) before it reaches deduplication and any sink
(console, file, BetterStack/Postgres). No sink ever sees raw PII, and all future
call sites are covered automatically.

Default patterns (applied in order; ordering prevents the patterns from eating
each other's digits):

| Pattern | Matches | Replacement |
|---------|---------|-------------|
| email | `user@host.tld` | `<email>` |
| ipv4 | four dotted octets, optional `:port` (incl. media IPs) | `<ip>` |
| phone (E.164) | `+` followed by 7-15 digits | `<phone>` |
| phone (id form) | 7-15 digits immediately followed by `@` (the `participant_id=<CID>@<ip>` / `conference_id=<DID>@<ip>` form) | `<phone>@` |

Bare in-sentence digit runs (no `+`, no trailing `@`) are deliberately **not**
redacted, so timestamps, `port=5060`, `samples=480`, byte/frame counters, UUIDs
and version strings are preserved. The structural fix for caller-ID embedded in
participant/conference IDs is CID-hashing (separate ticket).

### Toggle

Redaction is **enabled by default**. To disable it (intended for local
development only):

- Environment (read once at startup): `CORALIE_LOG_REDACT=0` (also accepts
  `false`, `no`, `off`, case-insensitive).
- Programmatically: `clog.SetRedactionEnabled(false)`.

### Custom patterns

Ops can replace the pattern set at runtime:

```go
r := clog.NewRedactor([]clog.RedactPattern{
    {Name: "email", Regex: `[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`, Replacement: "<email>"},
    {Name: "account", Regex: `ACC-\d{6}`, Replacement: "<account>"},
})
clog.SetRedactor(r)
```


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


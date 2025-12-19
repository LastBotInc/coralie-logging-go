# coralie-logging-go

A high-performance, feature-rich logging library for Go with deduplication, audio PCM/WAV logging, graceful shutdown, and comprehensive configuration.

## Quick Start

```go
package main

import (
    "context"
    "github.com/LastBotInc/coralie-logging-go/pkg/clog"
)

func main() {
    cfg := clog.DefaultConfig()
    cfg.Console.Enabled = true
    cfg.File.BaseDir = "./logs"
    
    clog.Init(cfg)
    defer clog.Shutdown(context.Background())
    
    clog.Info("Application", "Server starting on port 8080")
    clog.Success("Application", "Server started successfully")
    clog.Error("Application", "Connection failed: %v", err)
}
```

## Installation

```bash
go get github.com/LastBotInc/coralie-logging-go
```

## Features

- **Multiple log levels**: Debug, Info, Success, Warning, Fail, Error, Catastrophe
- **Deduplication**: Automatically collapses consecutive identical log lines
- **File routing**: Write different levels to different files
- **Console output**: Colorized, TTY-aware console logging
- **Audio logging**: Write PCM16 audio frames to WAV files
- **Graceful shutdown**: Drains queue, flushes all sinks, handles signals
- **Hooks**: Global and per-level hooks for custom processing
- **Performance**: Bounded queues, drop policies, minimal allocations

## Examples

See [EXAMPLES.md](EXAMPLES.md) for a complete list of examples.

Run the demo:
```bash
go run ./cmd/coralie-logging-demo
```

## Documentation

Comprehensive documentation is available in the [Documents/](Documents/) directory:

- [INDEX.md](Documents/INDEX.md) - Navigation guide
- [ARCHITECTURE.md](Documents/ARCHITECTURE.md) - System architecture
- [CONFIGURATION.md](Documents/CONFIGURATION.md) - Configuration options
- [LEVELS.md](Documents/LEVELS.md) - Log levels explained
- [DEDUPE.md](Documents/DEDUPE.md) - Deduplication behavior
- [AUDIO_PCM_WAV.md](Documents/AUDIO_PCM_WAV.md) - Audio logging guide
- [SHUTDOWN_PANIC_SIGNALS.md](Documents/SHUTDOWN_PANIC_SIGNALS.md) - Shutdown and signal handling
- [PERFORMANCE.md](Documents/PERFORMANCE.md) - Performance characteristics
- [TESTING.md](Documents/TESTING.md) - Testing guide
- [CHANGELOG.md](Documents/CHANGELOG.md) - Version history

## License

MIT License - see [LICENSE](LICENSE) file.


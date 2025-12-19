# Fyne Audio Monitor Example

This example demonstrates real-world integration of coralie-logging-go in a GUI application with audio capture and visualization.

## Features

- Fyne-based window application
- Microphone audio capture (with fake source for testing)
- FFT-based frequency domain visualization
- Interactive log level buttons
- Deduplication spam test button
- Optional audio WAV logging toggle

## Running

### Local Development

This example uses `replace` directive for local development:

```bash
cd examples/fyne-audio-monitor
go run .
```

### Production Usage

For production, remove the `replace` directive and use the published module:

```go
require github.com/LastBotInc/coralie-logging-go v1.0.0
```

## Testing

Tests use a fake audio source and do not require real hardware:

```bash
go test ./...
go test -race ./...
```

## Architecture

See [Documents/ARCHITECTURE.md](Documents/ARCHITECTURE.md) for detailed architecture documentation.


# Examples

This document lists all available examples and what they demonstrate.

## Demo CLI

**Location**: `cmd/coralie-logging-demo/`

A minimal non-GUI demonstration that proves:
- All log levels (Debug, Info, Success, Warning, Fail, Error, Catastrophe)
- Deduplication of consecutive identical messages
- File routing per level
- Synthetic PCM audio writing to WAV
- Clean shutdown

**Run**:
```bash
go run ./cmd/coralie-logging-demo
```

## Fyne Audio Monitor

**Location**: `examples/fyne-audio-monitor/`

A complete GUI application demonstrating real-world integration:
- Fyne-based window application
- Microphone audio capture
- FFT-based frequency domain visualization
- Interactive log level buttons
- Deduplication spam test button
- Optional audio WAV logging toggle
- Separate Go module proving importability

**Run**:
```bash
cd examples/fyne-audio-monitor
go run .
```

**Note**: This example uses `replace` directive for local development. See its README for production import instructions.


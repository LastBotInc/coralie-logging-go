# Audio PCM/WAV Logging

## Overview

coralie-logging-go can log PCM16 audio frames to WAV files, useful for debugging audio processing pipelines.

## Configuration

```go
cfg.Audio.Enabled = true
cfg.Audio.SampleRate = 44100
cfg.Audio.Channels = 1
cfg.Audio.BitsPerSample = 16
cfg.Audio.OutputDir = "./audio_logs"
cfg.Audio.FilenamePattern = "audio_%Y%m%d_%H%M%S.wav"
```

## Usage

Write PCM16 frames:

```go
// From int16 slice
frames := []int16{...}
clog.AudioWritePCM16(frames)

// From byte slice (little-endian PCM16)
data := []byte{...}
clog.AudioWriteBytesPCM16LE(data)
```

## File Format

- Format: WAV (RIFF)
- Sample format: PCM16 (signed 16-bit integers)
- Endianness: Little-endian
- Channels: Configurable (1 = mono, 2 = stereo)
- Sample rate: Configurable

## Buffering

Audio writes are buffered and flushed:
- On explicit flush
- On shutdown
- When buffer reaches a threshold

## Limitations

- Only PCM16 format is supported
- Files are created per session (new file on Init)
- No automatic file rotation (manual via shutdown/restart)

## Testing

Tests use synthetic audio data and do not require real hardware.


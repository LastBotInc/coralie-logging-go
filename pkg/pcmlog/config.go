// Package pcmlog: configuration for PCM/WAV logging.
package pcmlog

// Config holds configuration for PCM/WAV logging.
type Config struct {
	Enabled         bool
	SampleRate      int
	Channels        int
	BitsPerSample   int
	OutputDir       string
	FilenamePattern string
}


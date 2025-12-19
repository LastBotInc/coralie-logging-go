// Package clog: configuration structures.
package clog

// Config holds the complete configuration for the logger.
type Config struct {
	QueueSize  int
	DropPolicy string
	Console    ConsoleConfig
	File       FileConfig
	Dedupe     DedupeConfig
	Audio      AudioConfig
	Hooks      HooksConfig
}

// ConsoleConfig configures console output.
type ConsoleConfig struct {
	Enabled   bool
	Colors    bool
	OmitLevels map[Level]bool
}

// FileConfig configures file output.
type FileConfig struct {
	BaseDir  string
	PerLevel map[Level]string
}

// DedupeConfig configures deduplication.
type DedupeConfig struct {
	Enabled       bool
	SummaryFormat string
}

// AudioConfig configures audio PCM/WAV logging.
type AudioConfig struct {
	Enabled         bool
	SampleRate      int
	Channels        int
	BitsPerSample   int
	OutputDir       string
	FilenamePattern string
}

// HooksConfig configures hooks.
type HooksConfig struct {
	Global   []Hook
	PerLevel map[Level][]Hook
}

// DefaultConfig returns a default configuration.
func DefaultConfig() Config {
	return Config{
		QueueSize:  1000,
		DropPolicy: "drop_new",
		Console: ConsoleConfig{
			Enabled: true,
		},
		Dedupe: DedupeConfig{
			Enabled:       true,
			SummaryFormat: "last message repeated %d more times",
		},
	}
}


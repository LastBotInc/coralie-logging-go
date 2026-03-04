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
	// Sinks configures additional third-party sinks (e.g. BetterStack). Nil = no extra sinks.
	Sinks []SinkConfig
}

// SinkConfig configures one additional sink. Type determines which sink to use ("betterstack", etc.).
// MinLevel and OmitLevels apply level filtering for this sink. Format is "text" or "json".
// Type-specific fields: for Type "betterstack", set Token and optionally Endpoint.
type SinkConfig struct {
	Type       string       // "betterstack", etc.
	MinLevel   Level        // only emit events at or above this level; LevelDebug = all
	OmitLevels map[Level]bool
	Format     string       // "text" or "json"
	Token      string       // for betterstack: source token
	Endpoint   string       // for betterstack: ingest URL (default https://in.logs.betterstack.com)
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






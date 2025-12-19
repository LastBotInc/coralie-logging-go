// Package clog: console sink for log output.
package clog

import (
	"fmt"
	"os"

	"github.com/LastBotInc/coralie-logging-go/internal/term"
)

// consoleSink handles console output with optional colors and emojis.
type consoleSink struct {
	cfg      ConsoleConfig
	useColor bool
	useEmoji bool
}

// newConsoleSink creates a new console sink.
func newConsoleSink(cfg ConsoleConfig) *consoleSink {
	useColor := cfg.Colors
	if useColor {
		useColor = term.SupportsColor()
	}
	return &consoleSink{
		cfg:      cfg,
		useColor: useColor,
		useEmoji: useColor, // Use emojis when colors are enabled
	}
}

// write writes a formatted message to console.
func (s *consoleSink) write(level Level, iface, formatted string) {
	// Check if level should be omitted
	if s.cfg.OmitLevels != nil && s.cfg.OmitLevels[level] {
		return
	}

	levelStr := level.String()
	var output string

	if s.useColor || s.useEmoji {
		// Build colored/emoji output
		var prefix string
		if s.useEmoji {
			prefix = term.LevelEmoji(levelStr) + " "
		}
		if s.useColor {
			prefix += term.LevelColor(levelStr) + fmt.Sprintf("[%s]", levelStr) + term.ColorReset + " "
		} else if !s.useEmoji {
			prefix = fmt.Sprintf("[%s] ", levelStr)
		}
		output = prefix + iface + ": " + formatted
	} else {
		output = fmt.Sprintf("[%s] %s: %s", levelStr, iface, formatted)
	}

	fmt.Fprintln(os.Stdout, output)
}

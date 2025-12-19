// Package clog: console sink for log output.
package clog

import (
	"fmt"
	"os"
	"time"

	"github.com/LastBotInc/coralie-logging-go/internal/term"
	"github.com/LastBotInc/coralie-logging-go/internal/timefmt"
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
// Format: [<timestamp>][<emoji+level>][<facility>][<message>]
func (s *consoleSink) write(level Level, iface, formatted string) {
	// Check if level should be omitted
	if s.cfg.OmitLevels != nil && s.cfg.OmitLevels[level] {
		return
	}

	levelStr := level.String()
	now := time.Now()
	timestamp := timefmt.Format(now, "")

	var output string

	if s.useColor {
		// Format: [<timestamp>][<emoji+level>][<facility>]<message>
		// All brackets: dark gray
		bracketColor := term.ColorDarkGray
		
		// Timestamp: dark gray brackets
		timestampPart := bracketColor + "[" + term.ColorReset + timestamp + bracketColor + "]" + term.ColorReset

		// Emoji + Level: emoji + (optional space) + colored level string, dark gray brackets
		levelColor := term.LevelColor(levelStr)
		emojiLevelPart := ""
		if s.useEmoji {
			emoji := term.LevelEmoji(levelStr)
			// Add space for INFO, WARNING, DEBUG; no space for SUCCESS, FAIL, ERROR, CATASTROPHE
			space := " "
			if level == LevelSuccess || level == LevelFail || level == LevelError || level == LevelCatastrophe || level == LevelDebug {
				space = ""
			}
			emojiLevelPart = emoji + space + levelColor + levelStr + term.ColorReset
		} else {
			emojiLevelPart = levelColor + levelStr + term.ColorReset
		}
		emojiLevelPart = bracketColor + "[" + term.ColorReset + emojiLevelPart + bracketColor + "]" + term.ColorReset

		// Facility: bright white text, dark gray brackets
		facilityPart := bracketColor + "[" + term.ColorReset + term.ColorBrightWhite + iface + term.ColorReset + bracketColor + "]" + term.ColorReset

		// Message: light gray (default), green for success, yellow for warning, red for fail/error/catastrophe
		// No brackets around message
		messageColor := term.ColorLightGray
		switch level {
		case LevelSuccess:
			messageColor = term.ColorGreen
		case LevelWarning:
			messageColor = term.ColorYellow
		case LevelFail, LevelError, LevelCatastrophe:
			messageColor = term.ColorRed
		}
		messagePart := messageColor + formatted + term.ColorReset

		output = timestampPart + emojiLevelPart + facilityPart + messagePart
	} else {
		// No colors: plain format
		emojiLevelPart := ""
		if s.useEmoji {
			emoji := term.LevelEmoji(levelStr)
			// Add space for INFO, WARNING; no space for SUCCESS, FAIL, ERROR, CATASTROPHE, DEBUG
			space := " "
			if level == LevelSuccess || level == LevelFail || level == LevelError || level == LevelCatastrophe || level == LevelDebug {
				space = ""
			}
			emojiLevelPart = emoji + space + levelStr
		} else {
			emojiLevelPart = levelStr
		}
		output = fmt.Sprintf("[%s][%s][%s]%s", timestamp, emojiLevelPart, iface, formatted)
	}

	fmt.Fprintln(os.Stdout, output)
}

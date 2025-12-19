// Package term: color support detection and formatting.
package term

import "os"

// SupportsColor returns whether the terminal supports colors.
// Checks for TTY and common environment variables.
func SupportsColor() bool {
	if !IsTTY() {
		return false
	}
	// Check for NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	// Check for COLORTERM
	if os.Getenv("COLORTERM") != "" {
		return true
	}
	// Default to true for TTY
	return true
}

// Color codes for terminal output.
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorGray   = "\033[90m"
)

// LevelColor returns the color code for a log level.
func LevelColor(level string) string {
	switch level {
	case "DEBUG":
		return ColorGray
	case "INFO":
		return ColorBlue
	case "SUCCESS":
		return ColorGreen
	case "WARNING":
		return ColorYellow
	case "FAIL":
		return ColorMagenta
	case "ERROR":
		return ColorRed
	case "CATASTROPHE":
		return ColorRed
	default:
		return ColorReset
	}
}

// LevelEmoji returns the emoji for a log level.
func LevelEmoji(level string) string {
	switch level {
	case "DEBUG":
		return "🔍"
	case "INFO":
		return "ℹ️"
	case "SUCCESS":
		return "✅"
	case "WARNING":
		return "⚠️"
	case "FAIL":
		return "❌"
	case "ERROR":
		return "🔴"
	case "CATASTROPHE":
		return "💥"
	default:
		return ""
	}
}


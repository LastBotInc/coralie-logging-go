// Package timefmt provides time formatting utilities for log timestamps.
package timefmt

import "time"

// Format formats a time according to the given pattern.
// For now, uses a simple HH:MM:SS format.
func Format(t time.Time, pattern string) string {
	return t.Format("15:04:05")
}


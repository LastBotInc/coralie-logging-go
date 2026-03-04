// Package clog: log level definitions and constants.
package clog

// Level represents a log level.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelSuccess
	LevelWarning
	LevelFail
	LevelError
	LevelCatastrophe
)

// AtLeast reports whether l is at or above min in severity order
// (Debug < Info < ... < Catastrophe). Used for per-sink minimum level filtering.
func (l Level) AtLeast(min Level) bool {
	return l >= min
}

// String returns the string representation of the level.
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelSuccess:
		return "SUCCESS"
	case LevelWarning:
		return "WARNING"
	case LevelFail:
		return "FAIL"
	case LevelError:
		return "ERROR"
	case LevelCatastrophe:
		return "CATASTROPHE"
	default:
		return "UNKNOWN"
	}
}






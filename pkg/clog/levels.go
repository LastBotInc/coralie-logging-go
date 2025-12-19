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


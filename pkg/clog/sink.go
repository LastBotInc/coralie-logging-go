// Package clog: sink interface for log output.
package clog

// Sink writes log events to a destination (console, file, HTTP, etc.).
// The agent calls Write for each event; the sink may apply level filtering
// and formatting internally. Flush and Close are called during Shutdown;
// sinks that do not need them may use no-op implementations.
type Sink interface {
	Write(level Level, iface, formatted string)
	Flush()
	Close()
}

// levelFilter returns true if the event should be written given minLevel and omitSet.
// If minLevel is set (e.g. LevelInfo), only levels >= minLevel pass.
// If omitSet[level] is true, the event is dropped. OmitSet takes precedence.
func levelFilter(level, minLevel Level, omitSet map[Level]bool) bool {
	if omitSet != nil && omitSet[level] {
		return false
	}
	if minLevel > 0 && !level.AtLeast(minLevel) {
		return false
	}
	return true
}

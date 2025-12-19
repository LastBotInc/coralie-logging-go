// Package clog: event model for log entries.
package clog

// Event represents a log event.
type Event struct {
	Level   Level
	Iface   string
	Message string
	Params  []interface{}
}


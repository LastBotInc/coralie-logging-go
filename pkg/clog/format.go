// Package clog: formatters for log output (text, JSON).
package clog

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/LastBotInc/coralie-logging-go/internal/timefmt"
)

// Formatter formats a log event for a sink. All built-in sinks can use
// a formatter to change output format (e.g. text vs JSON).
type Formatter interface {
	Format(level Level, iface, formatted string, t time.Time) []byte
}

// TextFormatter produces human-readable lines: [timestamp][level][facility]message
// (same style as the file sink). No color.
type TextFormatter struct{}

// Format implements Formatter.
func (TextFormatter) Format(level Level, iface, formatted string, t time.Time) []byte {
	ts := timefmt.Format(t, "")
	return []byte(fmt.Sprintf("[%s][%s][%s]%s\n", ts, level.String(), iface, formatted))
}

// JSONFormatter produces one JSON object per event for machine consumption
// (e.g. BetterStack). Fields: dt (RFC3339), level, facility, message.
type JSONFormatter struct{}

type jsonEvent struct {
	Dt       string `json:"dt"`
	Level    string `json:"level"`
	Facility string `json:"facility"`
	Message  string `json:"message"`
}

// Format implements Formatter.
func (JSONFormatter) Format(level Level, iface, formatted string, t time.Time) []byte {
	ev := jsonEvent{
		Dt:       t.UTC().Format(time.RFC3339Nano),
		Level:    level.String(),
		Facility: iface,
		Message:  formatted,
	}
	b, _ := json.Marshal(ev)
	return append(b, '\n')
}

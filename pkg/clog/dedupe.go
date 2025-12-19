// Package clog: deduplication logic for consecutive identical messages.
package clog

import "fmt"

// dedupeState tracks deduplication state.
type dedupeState struct {
	lastLevel    Level
	lastIface    string
	lastMessage  string
	repeatCount  int
	enabled      bool
	summaryFormat string
}

// newDedupeState creates a new dedupe state.
func newDedupeState(cfg DedupeConfig) *dedupeState {
	return &dedupeState{
		enabled:       cfg.Enabled,
		summaryFormat: cfg.SummaryFormat,
	}
}

// check checks if an event is a duplicate and returns whether to suppress it.
// Returns (shouldSuppress, shouldEmitSummary).
func (d *dedupeState) check(level Level, iface, formatted string) (bool, bool) {
	if !d.enabled {
		return false, false
	}

	// Check if this matches the last message
	if d.lastLevel == level && d.lastIface == iface && d.lastMessage == formatted {
		d.repeatCount++
		return true, false // Suppress this message
	}

	// Different message - need to emit summary if there were repeats
	shouldEmitSummary := d.repeatCount > 0

	// Update state (but don't reset repeatCount yet - flushSummary will do that)
	d.lastLevel = level
	d.lastIface = iface
	d.lastMessage = formatted
	// Note: repeatCount will be reset by flushSummary() when summary is emitted

	return false, shouldEmitSummary
}

// flushSummary returns a summary message if there are pending repeats.
func (d *dedupeState) flushSummary() (Level, string, string, bool) {
	if !d.enabled || d.repeatCount == 0 {
		return 0, "", "", false
	}

	summary := fmt.Sprintf(d.summaryFormat, d.repeatCount)
	d.repeatCount = 0 // Reset after flushing

	return d.lastLevel, d.lastIface, summary, true
}

// reset resets the dedupe state.
func (d *dedupeState) reset() {
	d.lastLevel = 0
	d.lastIface = ""
	d.lastMessage = ""
	d.repeatCount = 0
}


// Package clog: statistics tracking.
package clog

import "sync/atomic"

// Stats holds logging statistics.
type Stats struct {
	DropsPerLevel map[Level]int64
	AcceptedCount int64
	EmittedCount  int64
}

// stats holds the global statistics.
var globalStats = struct {
	dropsPerLevel map[Level]*int64
	accepted      int64
	emitted       int64
}{
	dropsPerLevel: make(map[Level]*int64),
}

func init() {
	// Initialize atomic counters for each level
	for l := LevelDebug; l <= LevelCatastrophe; l++ {
		var v int64
		globalStats.dropsPerLevel[l] = &v
	}
}

// recordAccepted increments the accepted count.
func recordAccepted() {
	atomic.AddInt64(&globalStats.accepted, 1)
}

// recordEmitted increments the emitted count.
func recordEmitted() {
	atomic.AddInt64(&globalStats.emitted, 1)
}

// recordDrop increments the drop count for a level.
func recordDrop(level Level) {
	if counter, ok := globalStats.dropsPerLevel[level]; ok {
		atomic.AddInt64(counter, 1)
	}
}

// GetStats returns current statistics.
func GetStats() Stats {
	stats := Stats{
		DropsPerLevel: make(map[Level]int64),
		AcceptedCount: atomic.LoadInt64(&globalStats.accepted),
		EmittedCount:  atomic.LoadInt64(&globalStats.emitted),
	}
	for level, counter := range globalStats.dropsPerLevel {
		stats.DropsPerLevel[level] = atomic.LoadInt64(counter)
	}
	return stats
}


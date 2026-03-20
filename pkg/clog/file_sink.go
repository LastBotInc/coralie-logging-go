// Package clog: file sink for log output.
package clog

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/LastBotInc/coralie-logging-go/internal/timefmt"
)

// fileSink handles file output with per-level routing.
type fileSink struct {
	cfg      FileConfig
	files    map[Level]*os.File
	mu       sync.Mutex
}

// newFileSink creates a new file sink.
func newFileSink(cfg FileConfig) (*fileSink, error) {
	if cfg.BaseDir == "" {
		return nil, nil // File sink disabled
	}

	// Create base directory
	if err := os.MkdirAll(cfg.BaseDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	s := &fileSink{
		cfg:   cfg,
		files: make(map[Level]*os.File),
	}

	// Open files for configured levels
	if cfg.PerLevel != nil {
		for level, filename := range cfg.PerLevel {
			if filename == "" {
				continue // Skip empty filenames (omitted level)
			}
			filepath := filepath.Join(cfg.BaseDir, filename)
			file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600) //nolint:gosec // filepath is constructed from config, not user input
			if err != nil {
				// Close already opened files
				s.close()
				return nil, fmt.Errorf("failed to open log file %s: %w", filepath, err)
			}
			s.files[level] = file
		}
	}

	return s, nil
}

// write writes a formatted message to the appropriate file(s).
func (s *fileSink) write(level Level, iface, formatted string) {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	file, ok := s.files[level]
	if !ok {
		return // Level not configured for file output
	}

	// Format: [<timestamp>][<level>][<facility>]<message>
	now := time.Now()
	timestamp := timefmt.Format(now, "")
	output := fmt.Sprintf("[%s][%s][%s]%s\n", timestamp, level.String(), iface, formatted)
	_, _ = file.WriteString(output)
}

// flush flushes all open files.
func (s *fileSink) flush() {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, file := range s.files {
		_ = file.Sync()
	}
}

// close closes all open files.
func (s *fileSink) close() {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for level, file := range s.files {
		_ = file.Close()
		delete(s.files, level)
	}
}

// Write implements Sink. Writes a formatted message to the appropriate file(s).
func (s *fileSink) Write(level Level, iface, formatted string) {
	s.write(level, iface, formatted)
}

// Flush implements Sink. Syncs all open files.
func (s *fileSink) Flush() {
	s.flush()
}

// Close implements Sink. Closes all open files.
func (s *fileSink) Close() {
	s.close()
}

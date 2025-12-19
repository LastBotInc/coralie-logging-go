// Package pcmlog: WAV file writer for PCM16 frames.
package pcmlog

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Writer handles writing PCM16 frames to WAV files.
type Writer struct {
	cfg       Config
	file      *os.File
	dataSize  int64
	mu        sync.Mutex
	closed    bool
}

// NewWriter creates a new WAV file writer.
func NewWriter(cfg Config) (*Writer, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	// Create output directory
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create audio log directory: %w", err)
	}

	// Generate filename
	filename := cfg.FilenamePattern
	if filename == "" {
		filename = "audio_%Y%m%d_%H%M%S.wav"
	}
	
	// Simple time formatting (basic implementation)
	now := time.Now()
	filename = now.Format("20060102_150405.wav")
	if cfg.FilenamePattern != "" {
		// Try to parse pattern (simplified - just use timestamp)
		filename = fmt.Sprintf("audio_%s.wav", now.Format("20060102_150405"))
	}

	filepath := filepath.Join(cfg.OutputDir, filename)

	file, err := os.Create(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to create WAV file: %w", err)
	}

	w := &Writer{
		cfg:      cfg,
		file:     file,
		dataSize: 0,
	}

	// Write WAV header (will be updated on close)
	if err := w.writeHeader(); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to write WAV header: %w", err)
	}

	return w, nil
}

// writeHeader writes the WAV file header.
func (w *Writer) writeHeader() error {
	// RIFF header
	if _, err := w.file.WriteString("RIFF"); err != nil {
		return err
	}
	// File size (will be updated on close) - write placeholder
	if err := binary.Write(w.file, binary.LittleEndian, uint32(0)); err != nil {
		return err
	}
	// WAVE identifier
	if _, err := w.file.WriteString("WAVE"); err != nil {
		return err
	}

	// fmt chunk
	if _, err := w.file.WriteString("fmt "); err != nil {
		return err
	}
	// fmt chunk size
	if err := binary.Write(w.file, binary.LittleEndian, uint32(16)); err != nil {
		return err
	}
	// Audio format (1 = PCM)
	if err := binary.Write(w.file, binary.LittleEndian, uint16(1)); err != nil {
		return err
	}
	// Number of channels
	if err := binary.Write(w.file, binary.LittleEndian, uint16(w.cfg.Channels)); err != nil {
		return err
	}
	// Sample rate
	if err := binary.Write(w.file, binary.LittleEndian, uint32(w.cfg.SampleRate)); err != nil {
		return err
	}
	// Byte rate
	byteRate := uint32(w.cfg.SampleRate * w.cfg.Channels * w.cfg.BitsPerSample / 8)
	if err := binary.Write(w.file, binary.LittleEndian, byteRate); err != nil {
		return err
	}
	// Block align
	blockAlign := uint16(w.cfg.Channels * w.cfg.BitsPerSample / 8)
	if err := binary.Write(w.file, binary.LittleEndian, blockAlign); err != nil {
		return err
	}
	// Bits per sample
	if err := binary.Write(w.file, binary.LittleEndian, uint16(w.cfg.BitsPerSample)); err != nil {
		return err
	}

	// data chunk header
	if _, err := w.file.WriteString("data"); err != nil {
		return err
	}
	// data chunk size (will be updated on close) - write placeholder
	if err := binary.Write(w.file, binary.LittleEndian, uint32(0)); err != nil {
		return err
	}

	return nil
}

// updateHeader updates the WAV file header with correct sizes.
func (w *Writer) updateHeader() error {
	if w.file == nil {
		return nil
	}

	// Calculate file size (data size + header size)
	// Header is typically 44 bytes, but we'll calculate it
	headerSize := int64(44) // Standard WAV header size
	fileSize := headerSize + w.dataSize - 8 // -8 because RIFF size doesn't include first 8 bytes

	// Update RIFF chunk size (at offset 4)
	if _, err := w.file.Seek(4, 0); err != nil {
		return err
	}
	if err := binary.Write(w.file, binary.LittleEndian, uint32(fileSize)); err != nil {
		return err
	}

	// Update data chunk size (at offset 40, after "data" string)
	if _, err := w.file.Seek(40, 0); err != nil {
		return err
	}
	if err := binary.Write(w.file, binary.LittleEndian, uint32(w.dataSize)); err != nil {
		return err
	}

	return nil
}

// WritePCM16 writes PCM16 frames to the WAV file.
func (w *Writer) WritePCM16(frames []int16) error {
	if w == nil || w.closed {
		return nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return fmt.Errorf("writer is closed")
	}

	// Write frames as little-endian int16
	for _, frame := range frames {
		if err := binary.Write(w.file, binary.LittleEndian, frame); err != nil {
			return err
		}
		w.dataSize += 2 // int16 is 2 bytes
	}

	return nil
}

// WriteBytesPCM16LE writes PCM16 little-endian bytes to the WAV file.
func (w *Writer) WriteBytesPCM16LE(data []byte) error {
	if w == nil || w.closed {
		return nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return fmt.Errorf("writer is closed")
	}

	if _, err := w.file.Write(data); err != nil {
		return err
	}
	w.dataSize += int64(len(data))

	return nil
}

// Flush flushes buffered data to the WAV file.
func (w *Writer) Flush() error {
	if w == nil || w.closed {
		return nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		return w.file.Sync()
	}
	return nil
}

// Close closes the WAV file and updates the header.
func (w *Writer) Close() error {
	if w == nil || w.closed {
		return nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil
	}
	w.closed = true

	if w.file == nil {
		return nil
	}

	// Update header with correct sizes
	if err := w.updateHeader(); err != nil {
		w.file.Close()
		return err
	}

	return w.file.Close()
}

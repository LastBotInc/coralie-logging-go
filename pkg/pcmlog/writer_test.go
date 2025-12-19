package pcmlog

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAudioWriterCreatesWav(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := Config{
		Enabled:         true,
		SampleRate:      44100,
		Channels:        1,
		BitsPerSample:   16,
		OutputDir:       tmpDir,
		FilenamePattern: "test.wav",
	}

	writer, err := NewWriter(cfg)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	if writer == nil {
		t.Fatal("Writer is nil")
	}

	// Write some synthetic PCM16 data (sine wave samples)
	samples := make([]int16, 1000)
	for i := range samples {
		// Simple sine wave
		samples[i] = int16(1000 * (i % 100))
	}

	if err := writer.WritePCM16(samples); err != nil {
		t.Fatalf("Failed to write PCM16: %v", err)
	}

	if err := writer.Flush(); err != nil {
		t.Fatalf("Failed to flush: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close: %v", err)
	}

	// Verify file exists and is non-empty
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("No WAV file created")
	}

	wavFile := filepath.Join(tmpDir, files[0].Name())
	info, err := os.Stat(wavFile)
	if err != nil {
		t.Fatalf("Failed to stat WAV file: %v", err)
	}

	if info.Size() == 0 {
		t.Error("WAV file is empty")
	}

	if info.Size() < 1000 {
		t.Errorf("WAV file too small: %d bytes", info.Size())
	}
}

func TestAudioWriter_WriteBytesPCM16LE(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := Config{
		Enabled:       true,
		SampleRate:    44100,
		Channels:      1,
		BitsPerSample: 16,
		OutputDir:     tmpDir,
	}

	writer, err := NewWriter(cfg)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}

	// Write bytes directly
	data := make([]byte, 2000) // 1000 samples * 2 bytes
	for i := 0; i < 1000; i++ {
		// Write little-endian int16 values
		value := int16(i * 10)
		data[i*2] = byte(value)
		data[i*2+1] = byte(value >> 8)
	}

	if err := writer.WriteBytesPCM16LE(data); err != nil {
		t.Fatalf("Failed to write bytes: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close: %v", err)
	}

	// Verify file exists
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("No WAV file created")
	}
}

func TestAudioWriter_Disabled(t *testing.T) {
	cfg := Config{
		Enabled: false,
	}

	writer, err := NewWriter(cfg)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}

	if writer != nil {
		t.Error("Writer should be nil when disabled")
	}
}


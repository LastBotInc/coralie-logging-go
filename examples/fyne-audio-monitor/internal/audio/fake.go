// Package audio: fake audio source for testing.
package audio

import "math"

// FakeSource provides deterministic synthetic audio for testing.
type FakeSource struct {
	sampleRate int
	channels   int
	running    bool
	counter    int64
}

// NewFakeSource creates a new fake audio source.
func NewFakeSource(sampleRate, channels int) *FakeSource {
	return &FakeSource{
		sampleRate: sampleRate,
		channels:   channels,
	}
}

// Start starts the fake source.
func (f *FakeSource) Start() error {
	f.running = true
	return nil
}

// Stop stops the fake source.
func (f *FakeSource) Stop() error {
	f.running = false
	return nil
}

// Read returns synthetic audio data (sine wave at 440 Hz).
func (f *FakeSource) Read() ([]int16, error) {
	if !f.running {
		return nil, nil
	}

	// Generate 1024 samples
	samples := make([]int16, 1024*f.channels)
	for i := 0; i < 1024; i++ {
		// 440 Hz sine wave
		t := float64(f.counter+int64(i)) / float64(f.sampleRate)
		value := int16(1000 * math.Sin(2*math.Pi*440*t))
		for c := 0; c < f.channels; c++ {
			samples[i*f.channels+c] = value
		}
	}
	f.counter += 1024

	return samples, nil
}

// SampleRate returns the sample rate.
func (f *FakeSource) SampleRate() int {
	return f.sampleRate
}

// Channels returns the number of channels.
func (f *FakeSource) Channels() int {
	return f.channels
}


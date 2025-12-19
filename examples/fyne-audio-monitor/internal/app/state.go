// Package app manages application state.
package app

import (
	"sync"

	"github.com/LastBotInc/coralie-logging-go/examples/fyne-audio-monitor/internal/audio"
)

// State holds the application state.
type State struct {
	mu            sync.RWMutex
	audioSource   audio.AudioSource
	running        bool
	audioLogging   bool
	fftBins        []float64
}

// NewState creates a new application state.
func NewState() *State {
	return &State{
		fftBins: make([]float64, 64),
	}
}

// SetAudioSource sets the audio source.
func (s *State) SetAudioSource(source audio.AudioSource) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.audioSource = source
}

// GetAudioSource returns the audio source.
func (s *State) GetAudioSource() audio.AudioSource {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.audioSource
}

// SetRunning sets the running state.
func (s *State) SetRunning(running bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = running
}

// IsRunning returns whether the app is running.
func (s *State) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// SetAudioLogging sets audio logging state.
func (s *State) SetAudioLogging(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.audioLogging = enabled
}

// IsAudioLogging returns whether audio logging is enabled.
func (s *State) IsAudioLogging() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.audioLogging
}

// SetFFTBins sets the FFT bins.
func (s *State) SetFFTBins(bins []float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fftBins = bins
}

// GetFFTBins returns the FFT bins.
func (s *State) GetFFTBins() []float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]float64, len(s.fftBins))
	copy(result, s.fftBins)
	return result
}


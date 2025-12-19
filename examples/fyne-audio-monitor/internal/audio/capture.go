// Package audio provides audio capture interfaces and implementations.
package audio

// AudioSource is an interface for audio data sources.
type AudioSource interface {
	Start() error
	Stop() error
	Read() ([]int16, error)
	SampleRate() int
	Channels() int
}


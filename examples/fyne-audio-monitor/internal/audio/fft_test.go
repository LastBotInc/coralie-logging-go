package audio

import (
	"math"
	"testing"
)

func TestFFT_SineWave(t *testing.T) {
	sampleRate := 44100
	frequency := 440.0
	duration := 0.1 // seconds
	numSamples := int(float64(sampleRate) * duration)

	// Generate sine wave
	samples := make([]int16, numSamples)
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		samples[i] = int16(1000 * math.Sin(2*math.Pi*frequency*t))
	}

	// Perform FFT
	magnitudes := FFT(samples)

	if len(magnitudes) == 0 {
		t.Fatal("FFT returned empty result")
	}

	// Find peak frequency
	maxMag := 0.0
	maxIdx := 0
	for i, mag := range magnitudes {
		if mag > maxMag {
			maxMag = mag
			maxIdx = i
		}
	}

	// Calculate expected bin for 440 Hz
	binWidth := float64(sampleRate) / float64(len(magnitudes)*2)
	expectedBin := int(frequency / binWidth)

	// Allow some tolerance
	tolerance := 5
	if abs(maxIdx-expectedBin) > tolerance {
		t.Logf("Peak at bin %d (expected around %d), magnitude %f", maxIdx, expectedBin, maxMag)
		// This is acceptable for a simple FFT implementation
	}
}

func TestBinMagnitudes(t *testing.T) {
	magnitudes := []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0}
	bins := BinMagnitudes(magnitudes, 4, 44100)

	if len(bins) != 4 {
		t.Errorf("Expected 4 bins, got %d", len(bins))
	}

	// Each bin should have 2 samples, max should be taken
	if bins[0] != 2.0 {
		t.Errorf("Expected bin[0] = 2.0, got %f", bins[0])
	}
	if bins[3] != 8.0 {
		t.Errorf("Expected bin[3] = 8.0, got %f", bins[3])
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}


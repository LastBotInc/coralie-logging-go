// Package audio: FFT implementation for frequency domain analysis.
package audio

import "math"

// FFT performs a simple radix-2 FFT on real input data.
// Returns magnitude spectrum.
func FFT(samples []int16) []float64 {
	n := len(samples)
	if n == 0 {
		return nil
	}

	// Pad to next power of 2
	nextPow2 := 1
	for nextPow2 < n {
		nextPow2 <<= 1
	}

	// Convert to complex and pad
	complexData := make([]complex128, nextPow2)
	for i := 0; i < n; i++ {
		complexData[i] = complex(float64(samples[i]), 0)
	}

	// Perform FFT
	fftRadix2(complexData)

	// Calculate magnitudes (only first half, symmetric)
	magnitudes := make([]float64, nextPow2/2)
	for i := 0; i < nextPow2/2; i++ {
		magnitudes[i] = math.Sqrt(real(complexData[i])*real(complexData[i]) + imag(complexData[i])*imag(complexData[i]))
	}

	return magnitudes
}

// fftRadix2 performs in-place FFT using radix-2 algorithm.
func fftRadix2(data []complex128) {
	n := len(data)
	if n <= 1 {
		return
	}

	// Bit-reverse permutation
	j := 0
	for i := 1; i < n; i++ {
		bit := n >> 1
		for j&bit != 0 {
			j ^= bit
			bit >>= 1
		}
		j ^= bit
		if i < j {
			data[i], data[j] = data[j], data[i]
		}
	}

	// FFT
	for size := 2; size <= n; size <<= 1 {
		angle := -2 * math.Pi / float64(size)
		w := complex(math.Cos(angle), math.Sin(angle))
		for i := 0; i < n; i += size {
			wj := complex(1, 0)
			for j := 0; j < size/2; j++ {
				u := data[i+j]
				v := data[i+j+size/2] * wj
				data[i+j] = u + v
				data[i+j+size/2] = u - v
				wj *= w
			}
		}
	}
}

// BinMagnitudes bins FFT magnitudes into histogram bins.
func BinMagnitudes(magnitudes []float64, numBins int, sampleRate int) []float64 {
	if len(magnitudes) == 0 || numBins == 0 {
		return make([]float64, numBins)
	}

	bins := make([]float64, numBins)
	samplesPerBin := len(magnitudes) / numBins

	for i := 0; i < numBins; i++ {
		start := i * samplesPerBin
		end := start + samplesPerBin
		if end > len(magnitudes) {
			end = len(magnitudes)
		}

		max := 0.0
		for j := start; j < end; j++ {
			if magnitudes[j] > max {
				max = magnitudes[j]
			}
		}
		bins[i] = max
	}

	return bins
}


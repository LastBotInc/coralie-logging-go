// Package ui: compact histogram widget with 1px bars.
package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// HistogramWidget is a compact histogram with 1px bars.
type HistogramWidget struct {
	container fyne.CanvasObject
	bars       []*canvas.Rectangle
	maxHeight  float32
}

// NewHistogramWidget creates a new compact histogram widget.
func NewHistogramWidget(numBins int, width, height float32) *HistogramWidget {
	bars := make([]*canvas.Rectangle, numBins)
	for i := 0; i < numBins; i++ {
		bar := canvas.NewRectangle(color.RGBA{0, 200, 0, 255}) // Green bars
		bar.Resize(fyne.NewSize(1, 0))
		bars[i] = bar
	}

	// Create horizontal container for bars
	barContainer := container.NewWithoutLayout()
	for _, bar := range bars {
		barContainer.Add(bar)
	}

	// Position bars horizontally with 1px spacing (will be positioned from bottom on update)
	for i, bar := range bars {
		x := float32(i * 2) // 1px bar + 1px spacing
		bar.Move(fyne.NewPos(x, 0)) // Temporary position, will be updated
	}

	widget := &HistogramWidget{
		container: barContainer,
		bars:      bars,
		maxHeight: height,
	}

	return widget
}

// Update updates the histogram with new data.
// This must be called from the main Fyne thread using fyne.Do().
func (h *HistogramWidget) Update(data []float64) {
	if len(data) == 0 || len(h.bars) == 0 {
		return
	}

	// Find max value
	maxVal := 0.0
	for _, v := range data {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal == 0 {
		maxVal = 1
	}

	// Update bar heights
	for i, val := range data {
		if i >= len(h.bars) {
			break
		}
		bar := h.bars[i]
		progress := val / maxVal
		if progress > 1.0 {
			progress = 1.0
		}

		barHeight := float32(progress) * h.maxHeight
		if barHeight < 1 {
			barHeight = 1 // Minimum 1px for visibility
		}
		bar.Resize(fyne.NewSize(1, barHeight))
		// Position from bottom
		x := bar.Position().X
		y := h.maxHeight - barHeight
		bar.Move(fyne.NewPos(x, y))
		bar.Refresh()
	}
}

// Container returns the container widget.
func (h *HistogramWidget) Container() fyne.CanvasObject {
	return h.container
}


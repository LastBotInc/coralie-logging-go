// Package ui provides UI layout components.
package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// CreateHistogram creates a simple histogram visualization.
func CreateHistogram(data []float64) fyne.CanvasObject {
	if len(data) == 0 {
		return widget.NewLabel("No data")
	}

	// Create a simple bar chart using progress bars
	bars := make([]fyne.CanvasObject, len(data))
	maxVal := 0.0
	for _, v := range data {
		if v > maxVal {
			maxVal = v
		}
	}

	if maxVal == 0 {
		maxVal = 1
	}

	for i, val := range data {
		progress := val / maxVal
		if progress > 1.0 {
			progress = 1.0
		}
		bar := widget.NewProgressBar()
		bar.SetValue(progress)
		bar.Resize(fyne.NewSize(10, 100))
		bars[i] = bar
	}

	return container.NewGridWithColumns(len(bars), bars...)
}


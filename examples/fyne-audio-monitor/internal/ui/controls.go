// Package ui provides UI controls.
package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// CreateLogButtons creates buttons for each log level in a compact grid.
func CreateLogButtons(onLog func(level string)) *fyne.Container {
	levels := []string{"Debug", "Info", "Success", "Warning", "Fail", "Error", "Catastrophe"}
	
	buttons := make([]fyne.CanvasObject, len(levels))
	for i, level := range levels {
		level := level // capture
		btn := widget.NewButton(level, func() {
			onLog(level)
		})
		buttons[i] = btn
	}

	// Use 4 columns for compact layout
	return container.NewGridWithColumns(4, buttons...)
}

// CreateSpamButton creates a button to test deduplication.
func CreateSpamButton(onSpam func()) *widget.Button {
	return widget.NewButton("Spam 100x", onSpam)
}

// CreateAudioToggle creates a toggle for audio logging.
func CreateAudioToggle(onToggle func(bool)) *widget.Check {
	return widget.NewCheck("Audio WAV", onToggle)
}


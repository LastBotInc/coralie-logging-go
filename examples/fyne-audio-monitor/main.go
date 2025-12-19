// Fyne audio monitor example application.
package main

import (
	"context"
	"time"

	appstate "github.com/LastBotInc/coralie-logging-go/examples/fyne-audio-monitor/internal/app"
	"github.com/LastBotInc/coralie-logging-go/examples/fyne-audio-monitor/internal/audio"
	"github.com/LastBotInc/coralie-logging-go/examples/fyne-audio-monitor/internal/ui"
	"github.com/LastBotInc/coralie-logging-go/pkg/clog"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	// Initialize logging
	cfg := clog.DefaultConfig()
	cfg.Console.Enabled = true
	cfg.Console.Colors = true
	cfg.File.BaseDir = "./logs"
	cfg.File.PerLevel = map[clog.Level]string{
		clog.LevelError: "errors.log",
	}
	cfg.Dedupe.Enabled = true
	cfg.Audio.Enabled = true
	cfg.Audio.SampleRate = 44100
	cfg.Audio.Channels = 1
	cfg.Audio.OutputDir = "./audio_logs"

	clog.Init(cfg)
	defer clog.Shutdown(context.Background())

	stop := clog.InstallSignalHandler()
	defer stop()

	clog.Info("Application", "Starting audio monitor...")

	// Create application state
	state := appstate.NewState()

	// Use fake audio source (for demo - real implementation would use mic)
	audioSource := audio.NewFakeSource(44100, 1)
	state.SetAudioSource(audioSource)

	// Create Fyne app
	fyneApp := fyneapp.New()
	window := fyneApp.NewWindow("Audio Monitor")
	window.Resize(fyne.NewSize(400, 300))
	window.SetFixedSize(true)

	// Create compact histogram with 1px bars (64 bins, ~128px wide, 60px tall)
	histogramWidget := ui.NewHistogramWidget(64, 128, 60)

	logButtons := ui.CreateLogButtons(func(level string) {
		switch level {
		case "Debug":
			clog.Debug("UI", "Debug button clicked")
		case "Info":
			clog.Info("UI", "Info button clicked")
		case "Success":
			clog.Success("UI", "Success button clicked")
		case "Warning":
			clog.Warning("UI", "Warning button clicked")
		case "Fail":
			clog.Fail("UI", "Fail button clicked")
		case "Error":
			clog.Error("UI", "Error button clicked")
		case "Catastrophe":
			clog.Catastrophe("UI", "Catastrophe button clicked")
		}
	})

	spamButton := ui.CreateSpamButton(func() {
		for i := 0; i < 100; i++ {
			clog.Info("UI", "Spam message %d", i)
		}
	})

	audioToggle := ui.CreateAudioToggle(func(enabled bool) {
		state.SetAudioLogging(enabled)
		clog.Info("UI", "Audio logging: %v", enabled)
	})

	// Start/Stop button
	var startStopBtn *widget.Button
	startStopBtn = widget.NewButton("Start", func() {
		if state.IsRunning() {
			state.SetRunning(false)
			audioSource.Stop()
			startStopBtn.SetText("Start")
			clog.Info("Application", "Recording stopped")
		} else {
			state.SetRunning(true)
			audioSource.Start()
			startStopBtn.SetText("Stop")
			clog.Info("Application", "Recording started")
		}
	})

	// Compact layout
	content := container.NewVBox(
		widget.NewLabel("FFT"),
		histogramWidget.Container(),
		widget.NewSeparator(),
		logButtons,
		startStopBtn,
		container.NewHBox(spamButton, audioToggle),
	)

	window.SetContent(content)

	// Channel for FFT updates
	fftUpdates := make(chan []float64, 1)

	// Audio processing loop
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			if !state.IsRunning() {
				continue
			}

			samples, err := audioSource.Read()
			if err != nil || len(samples) == 0 {
				continue
			}

			// Perform FFT
			magnitudes := audio.FFT(samples)
			bins := audio.BinMagnitudes(magnitudes, 64, audioSource.SampleRate())
			state.SetFFTBins(bins)

			// Send update (non-blocking)
			select {
			case fftUpdates <- bins:
			default:
			}

			// Write to audio log if enabled
			if state.IsAudioLogging() {
				clog.AudioWritePCM16(samples)
			}
		}
	}()

	// UI update function - must run on main Fyne thread
	updateHistogram := func() {
		bins := state.GetFFTBins()
		if len(bins) == 0 {
			return
		}
		// Use fyne.Do to ensure we're on the main thread
		fyne.Do(func() {
			histogramWidget.Update(bins)
		})
	}

	// UI update loop - update from channel or timer
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				updateHistogram()
			case bins := <-fftUpdates:
				// Update state and UI
				state.SetFFTBins(bins)
				updateHistogram()
			}
		}
	}()

	// Show window
	window.ShowAndRun()

	clog.Info("Application", "Shutting down...")
}


package clog

import "testing"

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.QueueSize != 1000 {
		t.Errorf("DefaultConfig().QueueSize = %v, want 1000", cfg.QueueSize)
	}
	if cfg.DropPolicy != "drop_new" {
		t.Errorf("DefaultConfig().DropPolicy = %v, want drop_new", cfg.DropPolicy)
	}
	if !cfg.Console.Enabled {
		t.Error("DefaultConfig().Console.Enabled = false, want true")
	}
	if !cfg.Dedupe.Enabled {
		t.Error("DefaultConfig().Dedupe.Enabled = false, want true")
	}
}


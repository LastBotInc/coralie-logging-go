package clog

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestTextFormatter_Format(t *testing.T) {
	f := TextFormatter{}
	tm := time.Date(2025, 3, 4, 14, 5, 6, 0, time.UTC)
	b := f.Format(LevelInfo, "TestIface", "hello world", tm)
	s := string(b)
	if !strings.Contains(s, "INFO") {
		t.Errorf("expected INFO in output, got %q", s)
	}
	if !strings.Contains(s, "TestIface") {
		t.Errorf("expected TestIface in output, got %q", s)
	}
	if !strings.Contains(s, "hello world") {
		t.Errorf("expected message in output, got %q", s)
	}
	if !strings.HasSuffix(s, "\n") {
		t.Errorf("expected newline suffix, got %q", s)
	}
}

func TestJSONFormatter_Format(t *testing.T) {
	f := JSONFormatter{}
	tm := time.Date(2025, 3, 4, 14, 5, 6, 0, time.UTC)
	b := f.Format(LevelError, "Facility", "error message", tm)
	var ev jsonEvent
	if err := json.Unmarshal(b, &ev); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if ev.Level != "ERROR" {
		t.Errorf("Level = %q, want ERROR", ev.Level)
	}
	if ev.Facility != "Facility" {
		t.Errorf("Facility = %q, want Facility", ev.Facility)
	}
	if ev.Message != "error message" {
		t.Errorf("Message = %q, want error message", ev.Message)
	}
	if !strings.HasPrefix(ev.Dt, "2025-03-04") {
		t.Errorf("Dt = %q, want RFC3339 date", ev.Dt)
	}
}

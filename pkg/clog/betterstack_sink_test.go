package clog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestBetterStackSink_DisabledWithoutToken(t *testing.T) {
	s, err := newBetterStackSink(SinkConfig{Type: "betterstack"})
	if err != nil {
		t.Fatalf("newBetterStackSink: %v", err)
	}
	if s != nil {
		t.Error("expected nil sink when Token is empty")
	}
}

func TestBetterStackSink_Write_MinLevel(t *testing.T) {
	var mu sync.Mutex
	var received []betterstackEvent
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		var ev betterstackEvent
		if err := json.NewDecoder(r.Body).Decode(&ev); err != nil {
			t.Errorf("decode: %v", err)
			return
		}
		mu.Lock()
		received = append(received, ev)
		mu.Unlock()
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	s, err := newBetterStackSink(SinkConfig{
		Type:     "betterstack",
		Token:    "test-token",
		Endpoint: server.URL,
		MinLevel: LevelWarning,
	})
	if err != nil {
		t.Fatalf("newBetterStackSink: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil sink")
	}
	defer s.Close()

	s.Write(LevelInfo, "Iface", "info")     // below min: should not send
	s.Write(LevelWarning, "Iface", "warn")   // should send
	s.Write(LevelError, "Iface", "error")    // should send

	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	n := len(received)
	mu.Unlock()
	if n != 2 {
		t.Errorf("received %d events, want 2 (info filtered out)", n)
	}
}

func TestBetterStackSink_Write_OmitLevels(t *testing.T) {
	var mu sync.Mutex
	var received []betterstackEvent
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ev betterstackEvent
		if err := json.NewDecoder(r.Body).Decode(&ev); err != nil {
			t.Errorf("decode: %v", err)
			return
		}
		mu.Lock()
		received = append(received, ev)
		mu.Unlock()
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	s, err := newBetterStackSink(SinkConfig{
		Type:       "betterstack",
		Token:      "test-token",
		Endpoint:   server.URL,
		OmitLevels: map[Level]bool{LevelDebug: true, LevelInfo: true},
	})
	if err != nil {
		t.Fatalf("newBetterStackSink: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil sink")
	}
	defer s.Close()

	s.Write(LevelDebug, "Iface", "debug")  // omitted
	s.Write(LevelInfo, "Iface", "info")    // omitted
	s.Write(LevelWarning, "Iface", "warn") // sent

	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	n := len(received)
	mu.Unlock()
	if n != 1 {
		t.Errorf("received %d events, want 1", n)
	}
}

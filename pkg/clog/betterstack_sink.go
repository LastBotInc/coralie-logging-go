// Package clog: BetterStack (Logtail) HTTP sink for log output.
package clog

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

const defaultBetterStackEndpoint = "https://in.logs.betterstack.com"

// betterstackSink sends log events to BetterStack's HTTP ingest API.
type betterstackSink struct {
	client     *http.Client
	endpoint   string
	token      string
	minLevel   Level
	omitLevels map[Level]bool
	mu         sync.Mutex
	closed     bool
}

type betterstackEvent struct {
	Dt       string `json:"dt"`
	Level    string `json:"level"`
	Facility string `json:"facility"`
	Message  string `json:"message"`
	TraceID  string `json:"trace_id,omitempty"`
	SpanID   string `json:"span_id,omitempty"`
}

// newBetterStackSink creates a BetterStack sink from SinkConfig. Token must be set; Endpoint defaults if empty.
func newBetterStackSink(c SinkConfig) (*betterstackSink, error) {
	if c.Token == "" {
		return nil, nil // Disabled when no token
	}
	endpoint := c.Endpoint
	if endpoint == "" {
		endpoint = defaultBetterStackEndpoint
	}
	s := &betterstackSink{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		endpoint:   endpoint,
		token:      c.Token,
		minLevel:   c.MinLevel,
		omitLevels: c.OmitLevels,
	}
	return s, nil
}

// Write implements Sink. Sends one JSON event per call (no batching in v1).
func (s *betterstackSink) Write(level Level, iface, formatted string) {
	s.WriteEvent(Event{Level: level, Iface: iface, Message: formatted})
}

// WriteEvent implements EventSink. The agent has already formatted and redacted
// Message; correlation IDs are metadata and never alter the message body.
func (s *betterstackSink) WriteEvent(event Event) {
	if !levelFilter(event.Level, s.minLevel, s.omitLevels) {
		return
	}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	ev := betterstackEvent{
		Dt:       time.Now().UTC().Format(time.RFC3339Nano),
		Level:    event.Level.String(),
		Facility: event.Iface,
		Message:  event.Message,
		TraceID:  event.TraceID,
		SpanID:   event.SpanID,
	}
	body, _ := json.Marshal(ev)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, s.endpoint, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return
	}
	_ = resp.Body.Close()
}

// Flush implements Sink. No-op for v1 (no buffer).
func (s *betterstackSink) Flush() {}

// Close implements Sink. Marks sink closed so no further writes.
func (s *betterstackSink) Close() {
	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()
}

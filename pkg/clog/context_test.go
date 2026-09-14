// Package clog: correlation coverage through the public API and HTTP sink.
package clog

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestLogContextCorrelationAndRedaction(t *testing.T) {
	previous := RedactionEnabled()
	SetRedactionEnabled(true)
	t.Cleanup(func() { SetRedactionEnabled(previous) })
	requests := make(chan map[string]string, 6)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var event map[string]string
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			t.Errorf("decode log: %v", err)
		}
		requests <- event
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()
	hook := &testHook{}
	cfg := DefaultConfig()
	cfg.Console.Enabled = false
	cfg.Dedupe.Enabled = false
	cfg.Hooks.Global = []Hook{hook}
	cfg.Sinks = []SinkConfig{{Type: "betterstack", Token: "synthetic", Endpoint: server.URL}}
	Init(cfg)
	t.Cleanup(func() { Shutdown(context.Background()) })

	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{1}, SpanID: trace.SpanID{2}, TraceFlags: trace.FlagsSampled,
	})
	cases := []struct {
		name string
		ctx  context.Context
		want bool
	}{
		{"active", trace.ContextWithSpanContext(context.Background(), sc), true},
		{"unsampled", trace.ContextWithSpanContext(context.Background(), sc.WithTraceFlags(0)), true},
		{"background", context.Background(), false},
		{"nil", nil, false},
	}
	for _, tc := range cases {
		LogContext(tc.ctx, LevelInfo, "", "%s %s@%s", tc.name, "synthetic", "example.invalid")
	}
	Info("", "contextless")
	Shutdown(context.Background()) // drains the queue, without sleep/polling
	close(requests)
	events := hook.getEvents()
	if len(events) != 5 || len(requests) != 5 {
		t.Fatalf("hook/HTTP event counts = %d/%d, want 5/5", len(events), len(requests))
	}
	for i, event := range events {
		got := <-requests
		wantIDs := i < len(cases) && cases[i].want
		_, hasTrace := got["trace_id"]
		_, hasSpan := got["span_id"]
		if hasTrace != wantIDs || hasSpan != wantIDs {
			t.Errorf("event %d correlation fields = %v/%v, want %v", i, hasTrace, hasSpan, wantIDs)
		}
		if wantIDs && (event.TraceID != sc.TraceID().String() || event.SpanID != sc.SpanID().String()) {
			t.Errorf("event %d lost active span context", i)
		}
		if got["trace_id"] != event.TraceID || got["span_id"] != event.SpanID {
			t.Errorf("event %d hook/HTTP correlation differs", i)
		}
		if strings.Contains(event.Message, "synthetic@example.invalid") || event.Params != nil {
			t.Errorf("event %d retained unredacted content or params", i)
		}
		if got["message"] != event.Message || got["facility"] != defaultIface || got["level"] != "INFO" {
			t.Errorf("event %d changed message/level/facility semantics", i)
		}
		wantFields := 4
		if wantIDs {
			wantFields = 6
		}
		if len(got) != wantFields {
			t.Errorf("event %d JSON field count = %d, want %d", i, len(got), wantFields)
		}
	}
}

func TestLogContextDedupeDoesNotHideOtherSpans(t *testing.T) {
	hook := &testHook{}
	cfg := DefaultConfig()
	cfg.Console.Enabled = false
	cfg.Hooks.Global = []Hook{hook}
	Init(cfg)
	t.Cleanup(func() { Shutdown(context.Background()) })
	first := trace.NewSpanContext(trace.SpanContextConfig{TraceID: trace.TraceID{1}, SpanID: trace.SpanID{1}})
	second := first.WithSpanID(trace.SpanID{2})
	third := second.WithTraceID(trace.TraceID{2})
	for _, sc := range []trace.SpanContext{first, first, second, third} {
		ctx := trace.ContextWithSpanContext(context.Background(), sc)
		LogContext(ctx, LevelInfo, "call", "operation completed")
	}
	Shutdown(context.Background())
	var operations []Event
	for _, event := range hook.getEvents() {
		if event.Message == "operation completed" {
			operations = append(operations, event)
		}
	}
	if len(operations) != 3 {
		t.Fatalf("operation count = %d, want 3 distinct spans", len(operations))
	}
	if operations[0].SpanID == operations[1].SpanID || operations[1].TraceID == operations[2].TraceID {
		t.Fatal("correlation changed while the event was queued")
	}
}

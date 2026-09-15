// Package clog: event model for log entries.
package clog

import "go.opentelemetry.io/otel/trace"

// Event represents a log event.
type Event struct {
	Level   Level
	Iface   string
	Message string
	Params  []interface{}
	// TraceID and SpanID are populated for hooks and event-aware sinks when
	// LogContext receives a valid OpenTelemetry span context.
	TraceID string `json:"trace_id,omitempty"`
	SpanID  string `json:"span_id,omitempty"`

	// Keep enqueue allocation-free; hexadecimal formatting belongs to the agent.
	traceID trace.TraceID
	spanID  trace.SpanID
}

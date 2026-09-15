// Package clog: OpenTelemetry correlation without an SDK or exporter.
package clog

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

func (e *Event) captureContext(ctx context.Context) {
	if ctx == nil {
		return
	}
	sc := trace.SpanContextFromContext(ctx)
	if sc.IsValid() {
		e.traceID = sc.TraceID()
		e.spanID = sc.SpanID()
	}
}

func (e *Event) formatContext() {
	if e.traceID.IsValid() && e.spanID.IsValid() {
		e.TraceID = e.traceID.String()
		e.SpanID = e.spanID.String()
	}
}

// textMessage adds correlation to text sinks without changing the redacted
// Event.Message delivered to hooks or structured sinks. Contextless output is
// byte-for-byte the same as the existing Write path.
func (e Event) textMessage() string {
	if e.TraceID == "" || e.SpanID == "" {
		return e.Message
	}
	traceID, traceErr := trace.TraceIDFromHex(e.TraceID)
	spanID, spanErr := trace.SpanIDFromHex(e.SpanID)
	if traceErr != nil || spanErr != nil || !traceID.IsValid() || !spanID.IsValid() {
		return e.Message
	}
	return "[trace_id=" + e.TraceID + " span_id=" + e.SpanID + "] " + e.Message
}

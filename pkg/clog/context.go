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

package api

import (
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// annotateSpan attaches key/value pairs to the request's active span (created
// by the Trace middleware). Keys/values come as alternating string args so
// handlers can read like `annotateSpan(r, "request_uuid", id, "pii.type", t)`
// without importing the attribute package.
//
// No-op if there's no active span (e.g. in tests that bypass middleware).
// The keys we use here are the same ones the gateway's spans use, so a
// trace search in Tempo by `request_uuid=<X>` picks up gateway + tokenizer
// spans for the same prompt.
func annotateSpan(r *http.Request, kvs ...string) {
	span := trace.SpanFromContext(r.Context())
	if !span.IsRecording() {
		return
	}
	for i := 0; i+1 < len(kvs); i += 2 {
		span.SetAttributes(attribute.String(kvs[i], kvs[i+1]))
	}
}

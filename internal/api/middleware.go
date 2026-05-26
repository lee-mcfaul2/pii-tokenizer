package api

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/lee-mcfaul2/pii-tokenizer/internal/obs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			b := make([]byte, 8)
			_, _ = rand.Read(b)
			id = hex.EncodeToString(b)
		}
		w.Header().Set("X-Request-ID", id)
		r.Header.Set("X-Request-ID", id)
		next.ServeHTTP(w, r)
	})
}

func ContentTypeJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			ct := r.Header.Get("Content-Type")
			if ct != "" && ct != "application/json" && ct != "application/json; charset=utf-8" {
				WriteError(w, r, ErrSchemaValidationFailed, "content-type must be application/json")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic in handler",
						"err", rec,
						"stack", string(debug.Stack()),
						"path", r.URL.Path)
					WriteError(w, r, ErrInternal, "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func Trace(next http.Handler) http.Handler {
	tracer := otel.Tracer("pii-tokenizer")
	prop := otel.GetTextMapPropagator()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract W3C traceparent from inbound headers so this span links
		// to the caller's trace (typically agent-gateway) — without
		// extraction, every tokenizer request became its own root trace
		// and Tempo couldn't stitch the prompt lifecycle together.
		ctx := prop.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		ctx, span := tracer.Start(ctx, r.URL.Path)
		defer span.End()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		obs.RequestsTotal.WithLabelValues(r.URL.Path, strconv.Itoa(ww.Status())).Inc()
		obs.RequestDuration.WithLabelValues(r.URL.Path).Observe(time.Since(start).Seconds())
	})
}

func AccessLog(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", w.Header().Get("X-Request-ID"),
			)
		})
	}
}

package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lee-mcfaul2/pii-tokenizer/internal/scope"
)

type Server struct {
	scope  *scope.Service
	logger *slog.Logger
	Ready  func(ctx context.Context) error
}

func NewServer(s *scope.Service, logger *slog.Logger) *Server {
	return &Server{scope: s, logger: logger}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(RequestID)
	r.Use(Recover(s.logger))
	r.Use(AccessLog(s.logger))
	r.Use(ContentTypeJSON)

	r.Get("/healthz", s.handleHealthz)
	r.Get("/readyz", s.handleReadyz)
	r.Method("GET", "/metrics", promhttpHandler())

	r.Route("/v1", func(r chi.Router) {
		r.Post("/init_request", s.initHandler)
		r.Post("/tokenize", s.handleTokenize)
		r.Post("/detokenize", s.handleDetokenize)
		r.Post("/release_request", s.handleRelease)
	})

	return r
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 100*time.Millisecond)
	defer cancel()
	if s.Ready == nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	if err := s.Ready(ctx); err != nil {
		s.logger.Warn("not ready", "err", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleTokenize(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "TODO Task 18", http.StatusNotImplemented)
}
func (s *Server) handleDetokenize(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "TODO Task 18", http.StatusNotImplemented)
}
func (s *Server) handleRelease(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "TODO Task 19", http.StatusNotImplemented)
}

package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/lee-mcfaul2/pii-tokenizer/internal/api"
	"github.com/lee-mcfaul2/pii-tokenizer/internal/config"
	"github.com/lee-mcfaul2/pii-tokenizer/internal/kmaster"
	"github.com/lee-mcfaul2/pii-tokenizer/internal/obs"
	"github.com/lee-mcfaul2/pii-tokenizer/internal/scope"
	"github.com/lee-mcfaul2/pii-tokenizer/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("config load", "err", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownTracing, err := obs.SetupTracing(ctx, cfg.OTLPEndpoint, cfg.ServiceName)
	if err != nil {
		logger.Error("tracing setup", "err", err)
		os.Exit(1)
	}
	defer shutdownTracing(context.Background())

	rc, err := store.NewClient(cfg)
	if err != nil {
		logger.Error("redis client", "err", err)
		os.Exit(1)
	}
	defer rc.Close()

	km, err := kmaster.Open(ctx, cfg)
	if err != nil {
		logger.Error("kmaster open", "backend", cfg.KMasterBackend, "err", err)
		os.Exit(1)
	}
	defer km.Close()

	svc := scope.New(rc, km)

	srv := api.NewServer(svc, logger)
	srv.Ready = func(ctx context.Context) error {
		if err := rc.Ping(ctx); err != nil {
			return errors.New("redis: " + err.Error())
		}
		if _, err := km.CurrentVersion(ctx); err != nil {
			return errors.New("kmaster: " + err.Error())
		}
		return nil
	}

	httpSrv := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      srv.Router(),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	go func() {
		logger.Info("listening", "addr", cfg.ListenAddr, "backend", cfg.KMasterBackend, "redis_mode", cfg.RedisMode)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen", "err", err)
			cancel()
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	select {
	case sig := <-sigCh:
		logger.Info("shutdown signal received", "sig", sig.String())
	case <-ctx.Done():
		logger.Info("shutdown due to internal error")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown", "err", err)
	}
}

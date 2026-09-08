package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"yabbi_test/internal/auction"
	"yabbi_test/internal/dsp"
	"yabbi_test/internal/partner"
	httptransport "yabbi_test/transport/http"
)

const (
	dispatchTimeout = 200 * time.Millisecond

	addr            = ":8080"
	readTimeout     = 5 * time.Second
	writeTimeout    = 5 * time.Second
	shutdownTimeout = 5 * time.Second
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	store := partner.NewInMemoryStore()

	// В задании нет реальных DSP-эндпоинтов для интеграции — используем
	// dsp.FakeClient как продуктовую заглушку, а не как тестовый инструмент.
	// Реальная HTTP-реализация подключается через тот же интерфейс dsp.Client,
	// без изменений в auction/service.go — см. README.
	client := dsp.NewFakeClient()

	svc := auction.NewService(store, client, logger, dispatchTimeout)
	handler := httptransport.NewHandler(svc, logger)

	srv := &http.Server{
		Addr:         addr,
		Handler:      handler.Routes(),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("starting server", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped cleanly")
}

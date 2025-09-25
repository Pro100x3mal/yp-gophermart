package httpserver

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func StartServer(ctx context.Context, addr string, r *chi.Mux, logger *zap.Logger) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	logger.Info("starting server...", zap.String("address", addr))

	errCh := make(chan error, 1)
	go func() {
		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			errCh <- nil
			return
		}

		logger.Error("unexpected server error", zap.Error(err))
		errCh <- err
	}()

	select {
	case <-ctx.Done():
		logger.Info("server is shutting down...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("failed to shutdown server", zap.Error(err))
			return err
		}

		logger.Info("server shutdown complete")
		return nil
	case err := <-errCh:
		return err
	}
}

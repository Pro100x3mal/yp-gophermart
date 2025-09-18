package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/Pro100x3mal/yp-gophermart.git/internal/configs"
	"github.com/Pro100x3mal/yp-gophermart.git/internal/handlers"
	"github.com/Pro100x3mal/yp-gophermart.git/internal/infrastructure/httpserver"
	"github.com/Pro100x3mal/yp-gophermart.git/internal/infrastructure/logger"
	"github.com/Pro100x3mal/yp-gophermart.git/internal/repositories"
	"github.com/Pro100x3mal/yp-gophermart.git/internal/services"
	"go.uber.org/zap"
)

func main() {
	mainLogger := zap.NewExample()
	defer mainLogger.Sync()

	if err := run(); err != nil {
		mainLogger.Fatal("application failed:", zap.Error(err))
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg := configs.GetConfig()

	zLog, err := logger.NewLogger(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer zLog.Sync()

	dbLog := zLog.Named("database")
	srvLog := zLog.Named("server")
	httpLog := zLog.Named("http")

	repo, err := repositories.NewDB(ctx, cfg, dbLog)
	if err != nil {
		dbLog.Error("failed to initialize database storage", zap.Error(err))
		return err
	}
	defer repo.Close()

	authSvc := services.NewAuthService(repo, cfg)
	userHandler := handlers.NewUserHandler(authSvc, httpLog)
	router := handlers.NewRouter(userHandler)

	if err = httpserver.StartServer(ctx, cfg, router, srvLog); err != nil {
		srvLog.Error("server failed", zap.Error(err))
	}

	return err
}

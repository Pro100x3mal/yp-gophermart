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
	mainLog := zap.NewExample()
	defer mainLog.Sync()

	if err := run(); err != nil {
		mainLog.Fatal("application failed:", zap.Error(err))
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := configs.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	zLog, err := logger.NewLogger(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer zLog.Sync()

	dbLog := zLog.Named("database")
	httpLog := zLog.Named("http")
	srvLog := zLog.Named("server")

	repo, err := repositories.NewDB(ctx, cfg, dbLog)
	if err != nil {
		dbLog.Error("failed to initialize database storage", zap.Error(err))
		return err
	}
	defer repo.Close()

	jwtMgr := services.NewJWTManager(cfg)

	authSvc := services.NewAuthService(repo, jwtMgr)
	authH := handlers.NewAuthHandler(authSvc, httpLog)

	ordersSvc := services.NewOrdersService(repo)
	ordersH := handlers.NewOrdersHandler(ordersSvc, httpLog)

	router := handlers.NewRouter(httpLog, jwtMgr, authH, ordersH)

	if err = httpserver.StartServer(ctx, cfg, router, srvLog); err != nil {
		srvLog.Error("server failed", zap.Error(err))
	}

	return err
}

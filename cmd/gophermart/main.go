package main

import (
	"context"
	"fmt"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Pro100x3mal/yp-gophermart.git/internal/configs"
	"github.com/Pro100x3mal/yp-gophermart.git/internal/handlers"
	"github.com/Pro100x3mal/yp-gophermart.git/internal/infrastructure/httpclient"
	"github.com/Pro100x3mal/yp-gophermart.git/internal/infrastructure/httpserver"
	"github.com/Pro100x3mal/yp-gophermart.git/internal/infrastructure/jwtmanager"
	"github.com/Pro100x3mal/yp-gophermart.git/internal/infrastructure/logger"
	"github.com/Pro100x3mal/yp-gophermart.git/internal/infrastructure/worker"
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
	clientLog := zLog.Named("client")
	srvLog := zLog.Named("server")

	repo, err := repositories.NewDB(ctx, cfg, dbLog)
	if err != nil {
		dbLog.Error("failed to initialize database storage", zap.Error(err))
		return err
	}
	defer repo.Close()

	jwtMgr := jwtmanager.NewJWTManager(cfg.Secret, cfg.TokenTTL)

	authSvc := services.NewAuthService(repo, jwtMgr)
	authH := handlers.NewAuthHandler(authSvc, httpLog)

	ordersSvc := services.NewOrdersService(repo)
	ordersH := handlers.NewOrdersHandler(ordersSvc, httpLog)

	balanceSvc := services.NewBalanceService(repo)
	balanceH := handlers.NewBalanceHandler(balanceSvc, httpLog)

	router := handlers.NewRouter(httpLog, jwtMgr, authH, ordersH, balanceH)

	accrualClient := httpclient.NewAccrualClient(cfg.AccrualAddr)
	accrualSvc := services.NewAccrualService(accrualClient, repo, clientLog, cfg.BatchSize)

	wp := worker.NewWorkerPool(cfg.RateLimit)
	wp.Start()
	defer wp.Stop()

	go startAccrualPoller(ctx, wp, accrualSvc, cfg.PollInterval, clientLog)

	if err = httpserver.StartServer(ctx, cfg.RunAddr, router, srvLog); err != nil {
		srvLog.Error("server failed", zap.Error(err))
	}

	return err
}

func startAccrualPoller(ctx context.Context, wp *worker.WorkerPool, svc *services.AccrualService, d time.Duration, cLog *zap.Logger) {
	ticker := time.NewTicker(d)
	defer ticker.Stop()
	var (
		sleepUntil time.Time
		mu         sync.Mutex
	)

	for {
		select {
		case <-ctx.Done():
			cLog.Debug("stopping accrual poller")
			return
		case <-ticker.C:
			mu.Lock()
			shouldSleep := !sleepUntil.IsZero() && time.Now().Before(sleepUntil)
			mu.Unlock()
			if shouldSleep {
				continue
			}

			wp.Submit(func() {
				processed, retryAfter, err := svc.PollAndUpdate(ctx)
				if err != nil {
					cLog.Warn("accrual poll failed", zap.Error(err))
					return
				}
				if retryAfter > 0 {
					cLog.Debug("retrying accrual poll", zap.Duration("retry_after", retryAfter))
					mu.Lock()
					sleepUntil = time.Now().Add(retryAfter)
					mu.Unlock()
					return
				}
				if processed == 0 {
					cLog.Debug("no orders to process")
					return
				}
				cLog.Debug("accrual poll updated", zap.Int("processed", processed))
			})
		}
	}
}

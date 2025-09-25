package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Pro100x3mal/yp-gophermart.git/internal/models"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type AccrualClient interface {
	GetOrderAccrual(ctx context.Context, number string) (*models.AccrualResp, time.Duration, error)
}

type AccrualRepository interface {
	SelectOrdersForAccrualPollingTx(ctx context.Context, tx pgx.Tx, limit int) ([]models.Order, error)
	UpdateOrderStatusAndAccrual(ctx context.Context, accrualResp *models.AccrualResp) error
	MarkOrdersProcessingTx(ctx context.Context, tx pgx.Tx, ids []int64) error
	BeginTx(ctx context.Context) (pgx.Tx, error)
}

type AccrualService struct {
	client    AccrualClient
	repo      AccrualRepository
	logger    *zap.Logger
	batchSize int
}

func NewAccrualService(client AccrualClient, repo AccrualRepository, logger *zap.Logger, batchSize int) *AccrualService {
	if batchSize <= 0 {
		batchSize = 100
	}
	return &AccrualService{
		client:    client,
		repo:      repo,
		logger:    logger.With(zap.String("service", "accrual")),
		batchSize: batchSize,
	}
}

func (as *AccrualService) PollAndUpdate(ctx context.Context) (int, time.Duration, error) {
	tx, err := as.repo.BeginTx(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	orders, err := as.repo.SelectOrdersForAccrualPollingTx(ctx, tx, as.batchSize)
	if err != nil {
		return 0, 0, fmt.Errorf("error selecting orders: %w", err)
	}
	if len(orders) == 0 {
		return 0, 0, nil
	}

	ids := make([]int64, 0, len(orders))
	for _, order := range orders {
		if order.Status == models.StatusNew {
			ids = append(ids, order.ID)
		}
	}
	if len(ids) > 0 {
		if err = as.repo.MarkOrdersProcessingTx(ctx, tx, ids); err != nil {
			return 0, 0, fmt.Errorf("error processing orders: %w", err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return 0, 0, fmt.Errorf("error committing transaction: %w", err)
	}

	var processed int
	for _, order := range orders {
		select {
		case <-ctx.Done():
			return processed, 0, ctx.Err()
		default:
		}

		accrualResp, retryAfter, clErr := as.client.GetOrderAccrual(ctx, order.Number)
		if clErr != nil {
			if errors.Is(clErr, models.ErrAccrualOrderNotRegistered) {
				continue
			}
			if errors.Is(clErr, models.ErrAccrualOrderTooMany) {
				return processed, retryAfter, nil
			}
			as.logger.Error("accrual client error", zap.String("order", order.Number), zap.Error(clErr))
			continue
		}

		switch accrualResp.Status {
		case models.StatusRegistered, models.StatusProcessing:
			upd := &models.AccrualResp{
				Order:  order.Number,
				Status: models.StatusProcessing,
			}
			if err = as.repo.UpdateOrderStatusAndAccrual(ctx, upd); err != nil {
				as.logger.Error("failed to update order to PROCESSING", zap.String("order", order.Number), zap.Error(err))
				continue
			}
			processed++

		case models.StatusInvalid:
			upd := &models.AccrualResp{
				Order:  order.Number,
				Status: models.StatusInvalid,
			}
			if err = as.repo.UpdateOrderStatusAndAccrual(ctx, upd); err != nil {
				as.logger.Error("failed to update order to INVALID", zap.String("order", order.Number), zap.Error(err))
				continue
			}
			processed++

		case models.StatusProcessed:
			upd := &models.AccrualResp{
				Order:   order.Number,
				Status:  models.StatusProcessed,
				Accrual: accrualResp.Accrual,
			}
			if err = as.repo.UpdateOrderStatusAndAccrual(ctx, upd); err != nil {
				as.logger.Error("failed to update order to PROCESSED", zap.String("order", order.Number), zap.Error(err))
				continue
			}
			processed++

		default:
			as.logger.Error("unexpected accrual status", zap.String("order", order.Number), zap.String("status", accrualResp.Status))
			continue
		}
	}

	return processed, 0, nil
}

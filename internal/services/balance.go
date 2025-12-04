package services

import (
	"context"
	"fmt"

	"github.com/Pro100x3mal/yp-gophermart.git/internal/models"
	"github.com/jackc/pgx/v5"
)

type BalanceRepository interface {
	GetBalanceByUserID(ctx context.Context, userID int64) (*models.Balance, error)
	CreateWithdrawalTx(ctx context.Context, tx pgx.Tx, userID int64, wd *models.WithdrawReq) error
	GetListWithdrawals(ctx context.Context, userID int64) ([]models.Withdrawal, error)
	BeginTx(ctx context.Context) (pgx.Tx, error)
}

type BalanceService struct {
	repo BalanceRepository
}

func NewBalanceService(repo BalanceRepository) *BalanceService {
	return &BalanceService{
		repo: repo,
	}
}

func (bs *BalanceService) CalculateBalance(ctx context.Context, userID int64) (*models.Balance, error) {
	balance, err := bs.repo.GetBalanceByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate balance: %w", err)
	}
	return balance, nil
}

func (bs *BalanceService) WithdrawFunds(ctx context.Context, userID int64, wd *models.WithdrawReq) error {
	tx, err := bs.repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err = bs.repo.CreateWithdrawalTx(ctx, tx, userID, wd); err != nil {
		return fmt.Errorf("failed to insert withdrawal: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (bs *BalanceService) ListWithdrawals(ctx context.Context, userID int64) ([]models.Withdrawal, error) {
	list, err := bs.repo.GetListWithdrawals(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get list of withdrawals: %w", err)
	}
	return list, nil
}

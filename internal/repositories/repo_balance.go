package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/Pro100x3mal/yp-gophermart.git/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (db *DB) GetBalanceByUserID(ctx context.Context, userID int64) (*models.Balance, error) {
	query := `
		WITH
  			o AS (
    			SELECT COALESCE(SUM(accrual), 0) AS total_accrual
    			FROM orders
    			WHERE user_id = $1 AND status = 'PROCESSED'
  			),
  			w AS (
    			SELECT COALESCE(SUM("sum"), 0) AS total_withdrawn
    			FROM withdrawals
    			WHERE user_id = $1
  			)
		SELECT
  			(o.total_accrual - w.total_withdrawn) AS current,
  			w.total_withdrawn AS withdrawn
		FROM o, w;
`
	var balance models.Balance
	err := db.pool.QueryRow(ctx, query, userID).Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}
		return nil, fmt.Errorf("database error: failed to get balance by user_id: %w", err)
	}
	return &balance, nil
}

func (db *DB) CreateWithdrawalTx(ctx context.Context, tx pgx.Tx, userID int64, wd *models.WithdrawReq) error {
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, userID); err != nil {
		return fmt.Errorf("database error: failed to acquire advisory lock: %w", err)
	}

	qAccrual := `
		SELECT COALESCE(SUM(accrual), 0)
		FROM orders
		WHERE user_id = $1 AND status = 'PROCESSED'
	`
	qWithdrawn := `
		SELECT COALESCE(SUM("sum"), 0)
		FROM withdrawals
		WHERE user_id = $1
	`

	var totalAccrual, totalWithdrawn float64

	if err := tx.QueryRow(ctx, qAccrual, userID).Scan(&totalAccrual); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return fmt.Errorf("database error: failed to get accrual sum: %w", err)
	}
	if err := tx.QueryRow(ctx, qWithdrawn, userID).Scan(&totalWithdrawn); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return fmt.Errorf("database error: failed to get withdrawn sum: %w", err)
	}

	available := totalAccrual - totalWithdrawn
	if available < wd.Sum {
		return models.ErrPaymentRequired
	}

	qInsert := `
		INSERT INTO withdrawals (user_id, order_number, "sum")
		VALUES ($1, $2, $3)
	`
	if _, err := tx.Exec(ctx, qInsert, userID, wd.Order, wd.Sum); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return err
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return models.ErrWithdrawalOrderExists
		}
		return fmt.Errorf("database error: failed to insert withdrawal: %w", err)
	}

	return nil
}

func (db *DB) GetListWithdrawals(ctx context.Context, userID int64) ([]models.Withdrawal, error) {
	query := `
		SELECT order_number,sum,processed_at
		FROM withdrawals
		WHERE user_id = $1
		ORDER BY processed_at DESC
`
	rows, err := db.pool.Query(ctx, query, userID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}
		return nil, fmt.Errorf("database error: failed to get withdrawals: %w", err)
	}
	defer rows.Close()

	var withdrawals []models.Withdrawal
	for rows.Next() {
		var withdrawal models.Withdrawal
		if err = rows.Scan(&withdrawal.Order, &withdrawal.Sum, &withdrawal.ProcessedAt); err != nil {
			return nil, fmt.Errorf("database error: failed to scan withdrawal: %w", err)
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("database error: failed to iterate over withdrawals: %w", err)
	}
	return withdrawals, nil
}

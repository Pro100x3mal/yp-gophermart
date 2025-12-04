package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/Pro100x3mal/yp-gophermart.git/internal/models"
	"github.com/jackc/pgx/v5"
)

func (db *DB) SelectOrdersForAccrualPollingTx(ctx context.Context, tx pgx.Tx, limit int) ([]models.Order, error) {
	query := `
		SELECT id, user_id, number, status, accrual, uploaded_at
		FROM orders
		WHERE status IN ('NEW', 'PROCESSING')
		ORDER BY uploaded_at
		FOR UPDATE SKIP LOCKED
		LIMIT $1
`
	rows, err := tx.Query(ctx, query, limit)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}
		return nil, fmt.Errorf("database error: failed to list orders for accrual polling: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		if err = rows.Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			return nil, fmt.Errorf("database error: failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("database error: failed to iterate over orders: %w", err)
	}

	return orders, nil
}

func (db *DB) MarkOrdersProcessingTx(ctx context.Context, tx pgx.Tx, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	query := `UPDATE orders SET status = 'PROCESSING' WHERE id = ANY($1) AND status='NEW'`

	_, err := tx.Exec(ctx, query, ids)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return err
		}
		return fmt.Errorf("database error: failed to mark orders processing: %w", err)
	}

	return nil
}

func (db *DB) UpdateOrderStatusAndAccrual(ctx context.Context, accrualResp *models.AccrualResp) error {
	var (
		query string
		args  []any
	)

	if accrualResp.Status == models.StatusProcessed {
		query = `UPDATE orders SET status = $1, accrual = $2 WHERE number = $3`
		args = []any{accrualResp.Status, accrualResp.Accrual, accrualResp.Order}
	} else {
		query = `UPDATE orders SET status = $1 WHERE number = $2`
		args = []any{accrualResp.Status, accrualResp.Order}
	}

	ct, err := db.pool.Exec(ctx, query, args...)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return err
		}
		return fmt.Errorf("database error: failed to update order status and accrual: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

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

func (db *DB) InsertOrder(ctx context.Context, userID int64, num string) error {
	query := `
		INSERT INTO orders (user_id, number)
		VALUES ($1, $2)
	`
	_, err := db.pool.Exec(ctx, query, userID, num)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return err
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return models.ErrOrderExists
		}
		return fmt.Errorf("database error: failed to insert order: %w", err)
	}
	return nil
}

func (db *DB) GetOrderOwnerID(ctx context.Context, num string) (int64, error) {
	query := `
		SELECT user_id
		FROM orders
		WHERE number = $1
	`
	var id int64
	err := db.pool.QueryRow(ctx, query, num).Scan(&id)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return 0, err
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, models.ErrOrderNotFound
		}
		return 0, fmt.Errorf("database error: failed to get order owner id: %w", err)
	}
	return id, nil
}

func (db *DB) GetOrdersByUserID(ctx context.Context, userID int64) ([]models.Order, error) {
	query := `
		SELECT id, user_id, number, status, accrual, uploaded_at
		FROM orders
		WHERE user_id = $1
		ORDER BY uploaded_at DESC
	`
	rows, err := db.pool.Query(ctx, query, userID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}
		return nil, fmt.Errorf("database error: failed to list user orders: %w", err)
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

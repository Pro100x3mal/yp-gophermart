package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/Pro100x3mal/yp-gophermart.git/internal/models"
	"github.com/jackc/pgx/v5"
)

func (db *DB) CreateUser(ctx context.Context, login string, passHash []byte) (int64, error) {
	query := `
		INSERT INTO users (login, password_hash)
		VALUES ($1, $2)
		ON CONFLICT (login) DO NOTHING
		RETURNING id
`

	var id int64
	if err := db.pool.QueryRow(ctx, query, login, passHash).Scan(&id); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return 0, err
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, models.ErrUserAlreadyExists
		}
		return 0, fmt.Errorf("database error: failed to create user: %w", err)
	}

	return id, nil
}

func (db *DB) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	query := `
		SELECT id, login, password_hash, created_at
		FROM users
		WHERE login = $1
	`

	var u models.User
	if err := db.pool.QueryRow(ctx, query, login).Scan(&u.ID, &u.Login, &u.PasswordHash, &u.CreatedAt); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrUserNotFound
		}
		return nil, fmt.Errorf("database error: failed to get user: %w", err)
	}
	return &u, nil
}

func (db *DB) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	query := `
		SELECT id, login, password_hash, created_at
		FROM users
		WHERE id = $1
	`
	var u models.User
	if err := db.pool.QueryRow(ctx, query, id).Scan(&u.ID, &u.Login, &u.PasswordHash, &u.CreatedAt); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrUserNotFound
		}
		return nil, fmt.Errorf("database error: failed to get user: %w", err)
	}
	return &u, nil
}

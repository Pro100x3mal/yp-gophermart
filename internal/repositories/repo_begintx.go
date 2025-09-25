package repositories

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func (db *DB) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return db.pool.Begin(ctx)
}

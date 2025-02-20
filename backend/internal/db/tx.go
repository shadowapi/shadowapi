package db

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// InTx runs fn in a transaction.
func InTx[T any](ctx context.Context, dbp *pgxpool.Pool, fn func(tx pgx.Tx) (T, error)) (T, error) {
	var (
		rollback  = true
		resultset T
	)
	tx, err := dbp.Begin(ctx)
	if err != nil {
		slog.Error("begin transaction", "error", err)
		return resultset, err
	}
	defer func() {
		if rollback {
			if err = tx.Rollback(ctx); err != nil {
				slog.Error("rollback transaction", "error", err)
			}
		}
	}()

	resultset, err = fn(tx)
	if err != nil {
		return resultset, err
	}

	rollback = false
	if err := tx.Commit(ctx); err != nil {
		rollback = true
		slog.Error("commit transaction", "error", err)
		return resultset, err
	}

	return resultset, nil
}

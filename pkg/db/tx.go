package db

import (
	"context"
	"database/sql"

	"github.com/xmtp/example-notification-server-go/pkg/db/queries"
)

// RunInTx executes fn within a database transaction. If fn returns an error,
// the transaction is rolled back. Otherwise, it is committed.
func RunInTx(ctx context.Context, db *sql.DB, fn func(txq *queries.Queries) error) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	if err := fn(queries.New(tx)); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// RunInTxResult executes fn within a database transaction, returning both a
// value and an error. If fn returns an error, the transaction is rolled back.
// Otherwise, it is committed.
func RunInTxResult[T any](ctx context.Context, db *sql.DB, fn func(txq *queries.Queries) (T, error)) (T, error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		var zero T
		return zero, err
	}
	result, err := fn(queries.New(tx))
	if err != nil {
		_ = tx.Rollback()
		var zero T
		return zero, err
	}
	if err := tx.Commit(); err != nil {
		var zero T
		return zero, err
	}
	return result, nil
}

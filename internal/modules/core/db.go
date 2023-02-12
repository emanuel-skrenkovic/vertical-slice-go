package core

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type TransactionOption func(*sql.TxOptions)

func WithIsolationLevel(isolationLevel sql.IsolationLevel) TransactionOption {
	return func(opts *sql.TxOptions) {
		opts.Isolation = isolationLevel
	}
}

func Tx(
	ctx context.Context,
	db *sqlx.DB,
	transaction func(context.Context, *sqlx.Tx) error,
	opts ...TransactionOption,
) (err error) {
	options := sql.TxOptions{}

	for _, opt := range opts {
		opt(&options)
	}

	tx, err := db.BeginTxx(ctx, &options)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = errors.Wrapf(err, "%v", r)
			} else {
				err = fmt.Errorf("transaction panicked with: %v", r)
			}
		}
	}()

	err = transaction(ctx, tx)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Wrapf(err, "%w", rollbackErr)
		}

		return err
	}

	err = tx.Commit()
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Wrapf(err, "%w", rollbackErr)
		}

		return err
	}

	return err
}

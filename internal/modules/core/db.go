package core

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type TransactionOption func(*sql.TxOptions)

func WithIsolationLevel(isolationLevel sql.IsolationLevel) TransactionOption {
	return func(opts *sql.TxOptions) {
		opts.Isolation = isolationLevel
	}
}

func Tx(
	ctx context.Context,
	db *sql.DB,
	transaction func(context.Context, *sql.Tx) error,
	opts ...TransactionOption,
) (err error) {
	options := sql.TxOptions{}

	for _, opt := range opts {
		opt(&options)
	}

	tx, err := db.BeginTx(ctx, &options)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			err = errors.Join(fmt.Errorf("transaction panicked with: %v", r), err)

			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = errors.Join(rollbackErr, err)
			}
		}
	}()

	err = transaction(ctx, tx)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Join(rollbackErr, err)
		}

		return err
	}

	err = tx.Commit()
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Join(rollbackErr, err)
		}

		return err
	}

	return err
}

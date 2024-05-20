package dbstore

import (
	"context"
	"database/sql"
)

type DBQueryManager interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

type TxManager struct {
	DBQueryManager
	db *sql.DB
	tx *sql.Tx
}

func NewTxManager(db *sql.DB, tx *sql.Tx) TxManager {
	return TxManager{db: db, tx: tx}
}

func (t TxManager) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if t.tx == nil {
		return t.db.ExecContext(ctx, query, args...)
	} else {
		return t.tx.ExecContext(ctx, query, args...)
	}
}

func (t TxManager) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if t.tx == nil {
		return t.db.QueryContext(ctx, query, args...)
	} else {
		return t.tx.QueryContext(ctx, query, args...)
	}
}

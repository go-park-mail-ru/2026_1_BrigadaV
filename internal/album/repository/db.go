package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

type PgxPoolAdapter struct {
	Pool *pgxpool.Pool
}

func (a *PgxPoolAdapter) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return a.Pool.Query(ctx, sql, args...)
}

func (a *PgxPoolAdapter) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return a.Pool.QueryRow(ctx, sql, args...)
}

func (a *PgxPoolAdapter) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return a.Pool.Exec(ctx, sql, args...)
}

func (a *PgxPoolAdapter) Begin(ctx context.Context) (pgx.Tx, error) {
	return a.Pool.Begin(ctx)
}

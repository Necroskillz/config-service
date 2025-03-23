package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type UnitOfWorkRunner interface {
	Run(ctx context.Context, fn func(tx *Queries) error) error
}

type PgxUnitOfWorkRunner struct {
	db      *pgx.Conn
	queries *Queries
}

func NewPgxUnitOfWorkRunner(db *pgx.Conn, queries *Queries) UnitOfWorkRunner {
	return &PgxUnitOfWorkRunner{db: db, queries: queries}
}

func (u *PgxUnitOfWorkRunner) Run(ctx context.Context, fn func(tx *Queries) error) error {
	tx, err := u.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)
	qtx := u.queries.WithTx(tx)

	err = fn(qtx)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
}

func NewDB(url string) (*DB, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, url)

	if err != nil {
		return nil, err
	}

	// fail fast – ensure DSN is valid and reachable now
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pg ping: %w", err)
	}

	return &DB{pool: pool}, nil
}

func (s *DB) Close() {
	s.pool.Close()
}

func (s *DB) Pool() *pgxpool.Pool {
	return s.pool
}

func (s *DB) RunInTx(ctx context.Context, fn func(tx pgx.Tx) error) error {

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

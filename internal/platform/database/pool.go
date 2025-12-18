package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect establishes a pgx connection pool when a database URL is provided.
func Connect(ctx context.Context, url string) (*pgxpool.Pool, error) {
	if url == "" {
		return nil, nil
	}

	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}

	return pgxpool.NewWithConfig(ctx, cfg)
}

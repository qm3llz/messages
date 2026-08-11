package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDB(ctx context.Context) *pgxpool.Pool {
	dsn := os.Getenv("DATABASE_URL")

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		panic(err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		panic(err)
	}

	return pool
}

package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDB(dbURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		fmt.Println("something went wrong when creating db connection config")
		return nil, err
	}

	config.MaxConnIdleTime = 5 * time.Minute
	config.MaxConnLifetime = 30 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute
	config.MaxConns = 10
	config.MinConns = 2

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		fmt.Println("something went wrong when creating db connection pool")
		return nil, err
	}

	if err = pool.Ping(ctx); err != nil {
		fmt.Println("something went wrong when pinging db")
		return nil, err
	}

	fmt.Println("db connection successful")
	return pool, nil
}

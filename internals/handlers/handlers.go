package handlers

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/luponetn/hng-stage-1/internals/db"
)

type Handler struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewHandler(queries *db.Queries, pool *pgxpool.Pool) *Handler {
	return &Handler{
		queries: queries,
		pool:    pool,
	}
}

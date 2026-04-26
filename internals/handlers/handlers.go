package handlers

import (
	"github.com/luponetn/hng-stage-1/internals/db"
)

type Handler struct {
	queries *db.Queries
}

func NewHandler(queries *db.Queries) *Handler {
	return &Handler{
		queries: queries,
	}
}
package handler

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type SystemHandler struct {
	db     *sql.DB
	logger *zerolog.Logger
}

func (sh *SystemHandler) PingFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		err := sh.db.PingContext(ctx)
		if err == nil {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func NewSystemHandler(db *sql.DB, logger *zerolog.Logger) *SystemHandler {
	return &SystemHandler{
		db:     db,
		logger: logger,
	}
}

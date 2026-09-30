package handler

import (
	"database/sql"
	"net/http"

	"github.com/rs/zerolog"
)

type SystemHandler struct {
	db     *sql.DB
	logger *zerolog.Logger
}

func (sh *SystemHandler) PingFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := sh.db.PingContext(r.Context())
		if err != nil {
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

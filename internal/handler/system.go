package handler

import (
	"context"
	"go-metrics/internal/service"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type SystemHandler struct {
	pinger service.Pinger
	logger *zerolog.Logger
}

func (sh *SystemHandler) PingFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		err := sh.pinger.Ping(ctx)
		if err == nil {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func NewSystemHandler(pinger service.Pinger, logger *zerolog.Logger) *SystemHandler {
	return &SystemHandler{
		pinger: pinger,
		logger: logger,
	}
}

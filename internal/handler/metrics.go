package handler

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

func getMetricsParams(r *http.Request) (string, string, string) {
	metricType := chi.URLParam(r, "metricType")
	metricName := strings.TrimSpace(chi.URLParam(r, "metricName"))
	metricValue := chi.URLParam(r, "metricValue")

	return metricType, metricName, metricValue
}

func internalError(w http.ResponseWriter, logger *zerolog.Logger, msg string, err error) {
	logger.
		Error().
		Err(err).
		Msg(msg)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func badRequest(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

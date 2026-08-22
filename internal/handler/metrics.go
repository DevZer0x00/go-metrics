package handler

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

func checkMetricValue(w http.ResponseWriter, metricValue string) bool {
	if len(metricValue) == 0 {
		return false
	}

	return true
}

func getMetricsParams(r *http.Request) (string, string, string) {
	metricType := chi.URLParam(r, "metricType")
	metricName := strings.TrimSpace(chi.URLParam(r, "metricName"))
	metricValue := chi.URLParam(r, "metricValue")

	return metricType, metricName, metricValue
}

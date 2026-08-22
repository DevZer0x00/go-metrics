package handler

import (
	"go-metrics/internal/model"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

var allowedMetricTypes = map[string]bool{
	model.Counter: true,
	model.Gauge:   true,
}

func checkMetricType(w http.ResponseWriter, metricType string) bool {
	if _, exists := allowedMetricTypes[metricType]; !exists {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return false
	}

	return true
}

func checkMetricValue(w http.ResponseWriter, metricValue string) bool {
	if len(metricValue) == 0 {
		http.Error(w, "Bad request", http.StatusBadRequest)
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

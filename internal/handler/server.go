package handler

import (
	"go-metrics/internal/model"
	"go-metrics/internal/repository"
	"net/http"
	"strconv"
	"strings"
)

var allowedMetricTypes = map[string]bool{
	model.Counter: true,
	model.Gauge:   true,
}

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /update", metricUpdateHandler)
	mux.HandleFunc("POST /update/{path...}", metricUpdateHandler)

	return mux
}

func metricUpdateHandler(w http.ResponseWriter, r *http.Request) {
	metricType, metricName, metricValue := getMetricsParams(r)

	if len(metricName) == 0 {
		http.NotFound(w, r)
		return
	}

	if _, exists := allowedMetricTypes[metricType]; !exists || len(metricValue) == 0 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	switch metricType {
	case model.Counter:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		metrics, err := repository.Metrics.GetOrRegisterCounter(metricName)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		metrics.UpdateCounter(value)
		err = repository.Metrics.Save(metrics)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	case model.Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		metrics, err := repository.Metrics.GetOrRegisterGauge(metricName)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		metrics.UpdateGauge(value)
		err = repository.Metrics.Save(metrics)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
}

func getMetricsParams(r *http.Request) (string, string, string) {
	path := strings.Split(r.PathValue("path"), "/")

	metricType := path[0]

	metricName := ""
	if len(path) >= 2 {
		metricName = strings.TrimSpace(path[1])
	}

	metricValue := ""
	if len(path) >= 3 {
		metricValue = path[2]
	}

	return metricType, metricName, metricValue
}

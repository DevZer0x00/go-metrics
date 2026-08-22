package handler

import (
	"go-metrics/internal/model"
	"go-metrics/internal/repository"
	"net/http"
	"strconv"
)

func UpdateMetricHandler(w http.ResponseWriter, r *http.Request) {
	metricType, metricName, metricValue := getMetricsParams(r)

	if len(metricName) == 0 {
		http.NotFound(w, r)
		return
	}

	valid := checkMetricType(w, metricType) && checkMetricValue(w, metricValue)
	if !valid {
		return
	}

	metrics, err := repository.Metric.GetOrRegister(metricType, metricName)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	switch metricType {
	case model.Counter:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		metrics.UpdateDelta(value)
	case model.Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		metrics.UpdateValue(value)
		err = repository.Metric.Save(metrics)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	err = repository.Metric.Save(metrics)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

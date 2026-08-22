package handler

import (
	"fmt"
	"go-metrics/internal/model"
	"go-metrics/internal/repository"
	"net/http"
)

func GetMetricHandler(w http.ResponseWriter, r *http.Request) {
	metricType, metricName, _ := getMetricsParams(r)

	valid := checkMetricType(w, metricType)
	if !valid {
		return
	}

	if has, err := repository.Metric.Has(metricType, metricName); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	} else if !has {
		http.NotFound(w, r)
		return
	}

	metrics, _ := repository.Metric.GetOrRegister(metricType, metricName)

	var sValue string

	switch metricType {
	case model.Counter:
		sValue = fmt.Sprintf("%d", *metrics.Delta)
	case model.Gauge:
		sValue = fmt.Sprintf("%.3f", *metrics.Value)
	default:
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	_, err := w.Write([]byte(sValue))
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func GetAllMetricsHandler(w http.ResponseWriter, r *http.Request) {
	all, err := repository.Metric.All()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	for _, metric := range all {
		var sValue string

		switch metric.MType {
		case model.Counter:
			sValue = fmt.Sprintf("%d", *metric.Delta)
		case model.Gauge:
			sValue = fmt.Sprintf("%.3f", *metric.Value)
		}

		_, err = w.Write([]byte(sValue))
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

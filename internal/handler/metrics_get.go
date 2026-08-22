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

	switch metricType {
	case model.Counter:
		if has, err := repository.Metric.HasCounter(metricName); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		} else if !has {
			http.NotFound(w, r)
			return
		}

		metrics, err := repository.Metric.GetOrRegisterCounter(metricName)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		_, err = w.Write([]byte(fmt.Sprintf("%d", *metrics.Delta)))
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	case model.Gauge:
		if has, err := repository.Metric.HasGauge(metricName); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		} else if !has {
			http.NotFound(w, r)
			return
		}

		metrics, err := repository.Metric.GetOrRegisterGauge(metricName)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		_, err = w.Write([]byte(fmt.Sprintf("%.3f", *metrics.Value)))
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "Bad request", http.StatusBadRequest)
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
		switch metric.MType {
		case model.Counter:
			_, err = w.Write([]byte(fmt.Sprintf("%s %d\n", metric.ID, *metric.Delta)))
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		case model.Gauge:
			_, err = w.Write([]byte(fmt.Sprintf("%s %.3f\n", metric.ID, *metric.Value)))
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		}
	}

}

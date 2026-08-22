package handler

import (
	"errors"
	"go-metrics/internal/service"
	"log"
	"net/http"
)

type UpdateMetricsHandler struct {
	service *service.MetricsService
}

func (handler *UpdateMetricsHandler) HandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType, metricName, metricValue := getMetricsParams(r)

		if len(metricName) == 0 {
			http.NotFound(w, r)
			return
		}

		err := service.CheckMetricType(metricType)
		valid := err == nil && checkMetricValue(w, metricValue)
		if !valid {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		err = handler.service.UpdateFromStringValue(metricType, metricName, metricValue)
		if err != nil && errors.Is(err, service.ErrInvalidMetricValue) {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		} else if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func NewUpdateMetricsHandler(service *service.MetricsService) *UpdateMetricsHandler {
	return &UpdateMetricsHandler{
		service: service,
	}
}

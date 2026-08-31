package handler

import (
	"errors"
	"go-metrics/internal/assets"
	"go-metrics/internal/service"
	"net/http"

	"github.com/rs/zerolog/log"
)

type GetMetricsHandler struct {
	service *service.MetricsService
}

func (handler *GetMetricsHandler) GetHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType, metricName, _ := getMetricsParams(r)

		err := service.CheckMetricType(metricType)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		metric, err := handler.service.Get(metricType, metricName)
		if err != nil {
			if errors.Is(err, service.ErrMetricNotFound) {
				http.NotFound(w, r)
				return
			}

			log.
				Error().
				Err(err).
				Msg("get metric from repository")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		_, err = w.Write([]byte(metric.ValueToString()))
		if err != nil {
			log.
				Error().
				Err(err).
				Msg("write response")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func (handler *GetMetricsHandler) GetAllMetricsHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		all, err := handler.service.GetAll()
		if err != nil {
			log.
				Error().
				Err(err).
				Msg("write response")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		metricsData := make([]*assets.GetAlMetricsTemplateData, len(all))
		for index, metric := range all {
			metricsData[index] = &assets.GetAlMetricsTemplateData{
				Name:  metric.ID,
				Value: metric.ValueToString(),
			}
		}

		err = assets.ExecuteGetAllMetricsTemplate(w, metricsData)
		if err != nil {
			log.
				Error().
				Err(err).
				Msg("get template")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	}
}

func NewGetMetricsHandler(service *service.MetricsService) *GetMetricsHandler {
	return &GetMetricsHandler{
		service: service,
	}
}

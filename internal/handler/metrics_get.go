package handler

import (
	"encoding/json"
	"errors"
	"go-metrics/internal/assets"
	"go-metrics/internal/model"
	"go-metrics/internal/service"
	"io"
	"net/http"

	"github.com/rs/zerolog"
)

type GetMetricsHandler struct {
	service *service.MetricsService
	logger  *zerolog.Logger
}

func (handler *GetMetricsHandler) GetHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType, metricName, _ := getMetricsParams(r)

		err := service.CheckMetricType(metricType)
		if err != nil {
			badRequest(w)
			return
		}

		metric, err := handler.service.Get(metricType, metricName)
		if err != nil {
			if errors.Is(err, service.ErrMetricNotFound) {
				http.NotFound(w, r)
				return
			}

			internalError(w, handler.logger, "get metric from repository", err)
			return
		}

		_, err = w.Write([]byte(metric.ValueToString()))
		if err != nil {
			internalError(w, handler.logger, "write response", err)
			return
		}
	}
}

func (handler *GetMetricsHandler) GetAllMetricsHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		all, err := handler.service.GetAll()
		if err != nil {
			internalError(w, handler.logger, "write response", err)
			return
		}

		metricsData := make([]*assets.GetAlMetricsTemplateData, len(all))
		for index, metric := range all {
			metricsData[index] = &assets.GetAlMetricsTemplateData{
				Name:  metric.ID,
				Value: metric.ValueToString(),
			}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err = assets.ExecuteGetAllMetricsTemplate(w, metricsData)
		if err != nil {
			internalError(w, handler.logger, "get template", err)
			return
		}
	}
}

func (handler *GetMetricsHandler) GetMetricValueHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			internalError(w, handler.logger, "read body", err)
			return
		}

		metricReq := &model.Metric{}
		err = json.Unmarshal(body, &metricReq)
		if err != nil {
			handler.logger.Err(err).Msg("unmarshal metric request")
			badRequest(w)
			return
		}

		metric, err := handler.service.Get(metricReq.MType, metricReq.ID)
		if err != nil {
			if errors.Is(err, service.ErrMetricNotFound) {
				http.NotFound(w, r)
				return
			}

			internalError(w, handler.logger, "get metric from repository", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(metric)
		if err != nil {
			internalError(w, handler.logger, "json encode error", err)
			return
		}
	}
}

func NewGetMetricsHandler(service *service.MetricsService, logger *zerolog.Logger) *GetMetricsHandler {
	return &GetMetricsHandler{
		service: service,
		logger:  logger,
	}
}

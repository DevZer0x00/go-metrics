package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-metrics/internal/assets"
	"go-metrics/internal/model"
	"go-metrics/internal/service"
	"io"
	"net/http"
)

type GetMetricsHandler struct {
	service *service.MetricsService
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

			internalError(w, "get metric from repository", err)
			return
		}

		_, err = w.Write([]byte(metric.ValueToString()))
		if err != nil {
			internalError(w, "write response", err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (handler *GetMetricsHandler) GetAllMetricsHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		all, err := handler.service.GetAll()
		if err != nil {
			internalError(w, "write response", err)
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
			internalError(w, "get template", err)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	}
}

func (handler *GetMetricsHandler) GetMetricValueHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			badRequest(w)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			internalError(w, "read body", err)
			return
		}

		metricReq := &model.Metric{}
		err = json.Unmarshal(body, &metricReq)
		if err != nil {
			fmt.Println(err)
			badRequest(w)
			return
		}

		metric, err := handler.service.Get(metricReq.MType, metricReq.ID)
		if err != nil {
			if errors.Is(err, service.ErrMetricNotFound) {
				http.NotFound(w, r)
				return
			}

			internalError(w, "get metric from repository", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(metric)
		if err != nil {
			internalError(w, "json encode error", err)
			return
		}
	}
}

func NewGetMetricsHandler(service *service.MetricsService) *GetMetricsHandler {
	return &GetMetricsHandler{
		service: service,
	}
}

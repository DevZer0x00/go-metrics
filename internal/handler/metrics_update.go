package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-metrics/internal/model"
	"go-metrics/internal/service"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type UpdateMetricsHandler struct {
	service *service.MetricsService
}

func (handler *UpdateMetricsHandler) UpdateFromPathHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType, metricName, metricValue := getMetricsParams(r)

		err := handler.service.UpdateFromStringValue(metricType, metricName, metricValue)
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			for _, validationError := range validationErrors {
				switch validationError.Field() {
				case "ID":
					http.NotFound(w, r)
					return
				default:
					badRequest(w)
					return
				}
			}
		} else if errors.Is(err, service.ErrInvalidMetricValue) {
			badRequest(w)
		} else if err != nil {
			internalError(w, "update metrics", err)
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (handler *UpdateMetricsHandler) UpdateFromJSONHandlerFunc() http.HandlerFunc {
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

		metric := &model.Metric{}
		err = json.Unmarshal(body, &metric)
		if err != nil {
			fmt.Println(err)
			badRequest(w)
			return
		}

		err = handler.service.UpdateFromModel(metric)
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			for _, validationError := range validationErrors {
				switch validationError.Field() {
				case "ID":
					http.NotFound(w, r)
					return
				default:
					badRequest(w)
					return
				}
			}
		} else if err != nil {
			internalError(w, "update metrics", err)
		}

		w.WriteHeader(http.StatusOK)
	}
}

func NewUpdateMetricsHandler(service *service.MetricsService) *UpdateMetricsHandler {
	return &UpdateMetricsHandler{
		service: service,
	}
}

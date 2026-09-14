package handler

import (
	"encoding/json"
	"errors"
	"go-metrics/internal/model"
	"go-metrics/internal/service"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
)

type UpdateMetricsHandler struct {
	service *service.MetricsService
	logger  *zerolog.Logger
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
			internalError(w, handler.logger, "update metrics", err)
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (handler *UpdateMetricsHandler) UpdateFromJSONHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			internalError(w, handler.logger, "read body", err)
			return
		}

		metric := &model.Metric{}
		err = json.Unmarshal(body, &metric)
		if err != nil {
			handler.logger.Err(err).Msg("unmarshal metric request")
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
			internalError(w, handler.logger, "update metrics", err)
		}

		w.WriteHeader(http.StatusOK)
	}
}

func NewUpdateMetricsHandler(service *service.MetricsService, logger *zerolog.Logger) *UpdateMetricsHandler {
	return &UpdateMetricsHandler{
		service: service,
		logger:  logger,
	}
}

package routes

import (
	"go-metrics/internal/handler"
	"go-metrics/internal/service"

	"github.com/go-chi/chi/v5"
)

func NewRouter(metricsService *service.MetricsService) *chi.Mux {
	updateHandler := handler.NewUpdateMetricsHandler(metricsService)
	getHandler := handler.NewGetMetricsHandler(metricsService)

	router := chi.NewRouter()

	router.Post("/update/{metricType}/{metricName}/{metricValue}", updateHandler.HandlerFunc())
	router.Get("/value/{metricType}/{metricName}", getHandler.GetHandlerFunc())
	router.Get("/", getHandler.GetAllMetricsHandlerFunc())

	return router
}

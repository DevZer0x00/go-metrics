package routes

import (
	"go-metrics/internal/handler"
	"go-metrics/internal/http/middleware"
	"go-metrics/internal/service"

	"github.com/go-chi/chi/v5"
	coreMiddleware "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(metricsService *service.MetricsService) *chi.Mux {
	updateHandler := handler.NewUpdateMetricsHandler(metricsService)
	getHandler := handler.NewGetMetricsHandler(metricsService)

	router := chi.NewRouter()
	router.Use(
		middleware.GzipEncodedMiddleware(),
		middleware.LoggingMiddleware(),
		coreMiddleware.Compress(5, "text/html", "application/json"),
	)

	router.Post("/update/{metricType}/{metricName}/{metricValue}", updateHandler.UpdateFromPathHandlerFunc())
	router.Post("/update", updateHandler.UpdateFromJSONHandlerFunc())
	router.Post("/update/", updateHandler.UpdateFromJSONHandlerFunc())

	router.Get("/value/{metricType}/{metricName}", getHandler.GetHandlerFunc())
	router.Post("/value", getHandler.GetMetricValueHandler())
	router.Post("/value/", getHandler.GetMetricValueHandler())
	router.Get("/", getHandler.GetAllMetricsHandlerFunc())

	return router
}

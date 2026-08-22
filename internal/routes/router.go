package routes

import (
	"go-metrics/internal/handler"

	"github.com/go-chi/chi/v5"
)

func NewRouter() *chi.Mux {
	router := chi.NewRouter()
	router.Post("/update/{metricType}/{metricName}/{metricValue}", handler.UpdateMetricHandler)
	router.Get("/value/{metricType}/{metricName}", handler.GetMetricHandler)
	router.Get("/", handler.GetAllMetricsHandler)

	return router
}

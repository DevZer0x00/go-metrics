package main

import (
	"go-metrics/internal/config"
	"go-metrics/internal/repository"
	"go-metrics/internal/routes"
	"go-metrics/internal/service"
	"log"
	"net/http"
	"os"
)

func main() {
	cfg, err := config.ParseServerOptions(os.Environ(), os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	metricsService := service.NewMetricsService(repository.NewMemStorage())
	router := routes.NewRouter(metricsService)

	if err := http.ListenAndServe(cfg.Addr.Addr, router); err != nil {
		log.Fatal(err)
	}
}

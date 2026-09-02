package main

import (
	"go-metrics/internal/config"
	"go-metrics/internal/repository"
	"go-metrics/internal/routes"
	"go-metrics/internal/service"
	"net/http"
	"os"

	"github.com/rs/zerolog/log"
)

func main() {
	config.InitLog(os.Stdout)

	cfg, err := config.ParseServerOptions(os.Environ(), os.Args[1:])
	if err != nil {
		log.
			Fatal().
			Err(err).
			Msg("failed to parse server options")
	}

	metricsService := service.NewMetricsService(repository.NewMemStorage())
	router := routes.NewRouter(metricsService)

	if err = http.ListenAndServe(cfg.Addr.Addr, router); err != nil {
		log.
			Fatal().
			Err(err).
			Msg("failed to start server")
	}
}

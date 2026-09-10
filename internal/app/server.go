package app

import (
	"fmt"
	"go-metrics/internal/config"
	"go-metrics/internal/repository"
	"go-metrics/internal/routes"
	"go-metrics/internal/service"
	"net/http"

	"github.com/rs/zerolog"
)

func RunServer(environments []string, arguments []string, logger *zerolog.Logger) error {
	cfg, err := config.ParseServerOptions(environments, arguments)
	if err != nil {
		return fmt.Errorf("parse server options: %w", err)
	}

	logger.
		Info().
		Bool("restoreOnStart", cfg.Persistence.Storage.RestoreOnStart).
		Str("storageFilePath", cfg.Persistence.Storage.StorageFilePath).
		Msg("starting server")

	storage := repository.NewMemStorage()
	memStoragePersister := service.NewMetricsPersister(
		storage,
		logger,
		cfg.Persistence.Storage.RestoreOnStart,
		cfg.Persistence.Storage.StorageFilePath,
	)

	metricsService := service.NewMetricsService(storage, memStoragePersister, logger)
	err = metricsService.InitPersister(cfg.Persistence.Interval)
	if err != nil {
		return err
	}
	defer metricsService.Close()

	router := routes.NewRouter(metricsService, logger)

	return http.ListenAndServe(cfg.Addr.Addr, router)
}

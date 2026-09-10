package app

import (
	"fmt"
	"go-metrics/internal/config"
	"go-metrics/internal/repository"
	"go-metrics/internal/routes"
	"go-metrics/internal/service"
	"net/http"
	"time"

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
	err = memStoragePersister.Init()
	if err != nil {
		return err
	}

	ticker := time.NewTicker(time.Second * time.Duration(cfg.Persistence.Interval))
	defer ticker.Stop()

	flushFunc := func() {
		err := memStoragePersister.Flush()
		if err != nil {
			logger.Error().Err(err).Msg("failed to flush memory storage")
		}
	}

	go func() {
		for range ticker.C {
			flushFunc()
		}
	}()
	defer flushFunc()

	metricsService := service.NewMetricsService(storage)
	router := routes.NewRouter(metricsService, logger)

	return http.ListenAndServe(cfg.Addr.Addr, router)
}

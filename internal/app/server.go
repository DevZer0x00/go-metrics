package app

import (
	"fmt"
	"go-metrics/internal/config"
	"go-metrics/internal/repository"
	"go-metrics/internal/routes"
	"go-metrics/internal/service"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

func RunServer(environments []string, arguments []string) error {
	config.InitLog(os.Stdout)

	cfg, err := config.ParseServerOptions(environments, arguments)
	if err != nil {
		return fmt.Errorf("parse server options: %w", err)
	}

	log.
		Info().
		Bool("restoreOnStart", cfg.Persistence.Storage.RestoreOnStart).
		Str("storageFilePath", cfg.Persistence.Storage.StorageFilePath).
		Msg("starting server")

	storage := repository.NewMemStorage()
	memStoragePersister := service.NewMetricsPersister(
		storage,
		cfg.Persistence.Storage.RestoreOnStart,
		cfg.Persistence.Storage.StorageFilePath,
	)
	err = memStoragePersister.Init()
	if err != nil {
		log.Error().Err(err).Msg("failed to init memory storage")
	}

	ticker := time.NewTicker(time.Second * time.Duration(cfg.Persistence.Interval))
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			err := memStoragePersister.Flush()
			if err != nil {
				log.Error().Err(err).Msg("failed to flush memory storage")
			}
		}
	}()

	metricsService := service.NewMetricsService(storage)
	router := routes.NewRouter(metricsService)

	return http.ListenAndServe(cfg.Addr.Addr, router)
}

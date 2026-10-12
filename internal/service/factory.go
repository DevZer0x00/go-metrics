package service

import (
	"database/sql"
	"go-metrics/internal/config"
	"go-metrics/internal/repository"
	"strings"

	"github.com/rs/zerolog"
)

func MetricRepositoryFactory(db *sql.DB, cfg *config.ServerPersistence, logger *zerolog.Logger) MetricRepository {
	if len(strings.TrimSpace(cfg.Database.DSN)) > 0 {
		return repository.NewPostgresStorage(db, logger)
	}

	return repository.NewMemStorage()
}

func MetricRepositoryPersisterFactory(repo MetricRepository, cfg *config.ServerPersistence, logger *zerolog.Logger) MetricRepositoryPersister {
	if len(strings.TrimSpace(cfg.Database.DSN)) > 0 {
		return NewNopPersister()
	}

	if len(strings.TrimSpace(cfg.File.StorageFilePath)) == 0 {
		return NewNopPersister()
	}

	return NewMetricsPersister(
		repo.(*repository.MemStorage),
		logger,
		cfg.File.RestoreOnStart,
		cfg.File.StorageFilePath,
	)
}

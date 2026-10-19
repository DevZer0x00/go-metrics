package service

import (
	"database/sql"
	"go-metrics/internal/config"
	"go-metrics/internal/repository"
	"strings"

	"github.com/rs/zerolog"
)

func MetricRepositoryFactory(db *sql.DB, cfg *config.ServerPersistence, logger *zerolog.Logger) (MetricRepository, error) {
	if len(strings.TrimSpace(cfg.Database.DSN)) > 0 {
		return repository.NewPostgresStorage(db, logger)
	}

	return repository.NewMemStorage(), nil
}

func MetricRepositoryPersisterFactory(repo MetricRepository, cfg *config.ServerPersistence, logger *zerolog.Logger) MetricRepositoryPersister {
	if len(strings.TrimSpace(cfg.Database.DSN)) > 0 {
		return NewNopPersister()
	}

	if len(strings.TrimSpace(cfg.File.StorageFilePath)) == 0 {
		return NewNopPersister()
	}

	repo, ok := repo.(*repository.MemStorage)
	if !ok {
		return NewNopPersister()
	}

	return NewMetricsPersister(
		repo.(*repository.MemStorage),
		logger,
		cfg.File.RestoreOnStart,
		cfg.File.StorageFilePath,
	)
}

func PingerFactory(db *sql.DB, cfg *config.ServerPersistence) Pinger {
	if len(strings.TrimSpace(cfg.Database.DSN)) > 0 {
		return NewDBPinger(db)
	}

	return &NopPinger{}
}

package app

import (
	"context"
	"database/sql"
	"fmt"
	"go-metrics/internal/config"
	"go-metrics/internal/repository"
	"go-metrics/internal/routes"
	"go-metrics/internal/service"
	"go-metrics/migrations"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"
)

func RunServer(environments []string, arguments []string, logger *zerolog.Logger) error {
	cfg, err := config.ParseServerOptions(environments, arguments)
	if err != nil {
		return fmt.Errorf("parse server options: %w", err)
	}

	logger.
		Info().
		Bool("restoreOnStart", cfg.Persistence.File.RestoreOnStart).
		Str("storageFilePath", cfg.Persistence.File.StorageFilePath).
		Str("databaseDSN", cfg.Persistence.Database.DSN).
		Msg("starting server")

	db, err := sql.Open("pgx", cfg.Persistence.Database.DSN)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx := context.Background()
	storage := service.MetricRepositoryFactory(ctx, db, cfg.Persistence)

	if _, ok := storage.(*repository.PostgresStorage); ok {
		err = migrations.RunMigrations(ctx, db)
		if err != nil {
			return err
		}
	}

	persister := service.MetricRepositoryPersisterFactory(storage, cfg.Persistence, logger)

	metricsService := service.NewMetricsService(storage, persister, logger)
	err = metricsService.InitPersister(cfg.Persistence.Interval)
	if err != nil {
		return err
	}
	defer metricsService.Close()

	router := routes.NewRouter(metricsService, db, logger)

	return http.ListenAndServe(cfg.Addr.Addr, router)
}

package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-metrics/internal/config"
	"go-metrics/internal/repository"
	"go-metrics/internal/routes"
	"go-metrics/internal/service"
	"go-metrics/migrations"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"
)

func RunServer(ctx context.Context, environments []string, arguments []string, logger *zerolog.Logger) error {
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

	storage := service.MetricRepositoryFactory(db, cfg.Persistence, logger)

	if _, ok := storage.(*repository.PostgresStorage); ok {
		err = migrations.RunMigrations(db)
		if err != nil {
			return err
		}
	}

	persister := service.MetricRepositoryPersisterFactory(storage, cfg.Persistence, logger)
	metricsService := service.NewMetricsService(storage, persister, logger)

	err = metricsService.InitPersister(ctx, cfg.Persistence.Interval)
	if err != nil {
		return err
	}
	defer metricsService.Close()

	router := routes.NewRouter(metricsService, db, logger)

	server := &http.Server{
		Addr:    cfg.Addr.Addr,
		Handler: router,
	}

	ch := make(chan error)

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			ch <- err
		}
	}()

	for {
		select {
		case <-ctx.Done():
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()

			if err = server.Shutdown(shutdownCtx); err != nil {
				logger.Error().Err(err).Msg("shutdown server")
			}

			return nil
		case err = <-ch:
			return err
		}
	}
}

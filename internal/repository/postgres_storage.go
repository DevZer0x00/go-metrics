package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-metrics/internal/helper"
	"go-metrics/internal/model"
	"go-metrics/migrations"
	"maps"
	"slices"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

type PostgresStorage struct {
	db              *sql.DB
	logger          *zerolog.Logger
	retryableHelper *helper.RetryableHelper
}

func (ps *PostgresStorage) GetOrRegister(mtype, id string) (*model.Metric, error) {
	row := ps.db.QueryRow(
		"SELECT delta, value FROM metric WHERE hash = $1 AND mtype = $2 ",
		model.GetMetricHash(id),
		mtype,
	)

	var (
		delta sql.NullInt64
		value sql.NullFloat64
	)

	err := row.Scan(&delta, &value)

	if errors.Is(err, sql.ErrNoRows) {
		return model.NewMetric(id, mtype), nil
	}

	metric := model.NewMetric(id, mtype)

	switch mtype {
	case model.Counter:
		metric.Delta = &delta.Int64
	case model.Gauge:
		metric.Value = &value.Float64
	}

	return metric, err
}

func (ps *PostgresStorage) All() ([]*model.Metric, error) {
	rows, err := ps.db.Query(
		"SELECT id, mtype, delta, value FROM metric ORDER BY mtype, hash",
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	models := make([]*model.Metric, 0)

	for rows.Next() {
		metric := &model.Metric{}
		err := rows.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
		if err != nil {
			return nil, err
		}

		models = append(models, metric)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return models, nil
}

func (ps *PostgresStorage) Has(mtype, name string) (bool, error) {
	row := ps.db.QueryRow(
		"SELECT 1 FROM metric WHERE hash = $1 AND mtype = $2 ",
		model.GetMetricHash(name),
		mtype,
	)

	if errors.Is(row.Scan(), sql.ErrNoRows) {
		return false, nil
	}

	return true, nil
}

func (ps *PostgresStorage) Save(ctx context.Context, metric *model.Metric) error {
	switch metric.MType {
	case model.Counter:
		return ps.retryableHelper.Run(ctx, func() error {
			_, err := ps.db.ExecContext(
				ctx,
				`INSERT INTO metric (id, mtype, delta, value, hash) VALUES ($1, $2, $3, $4, $5) 
			   ON CONFLICT (mtype, hash) DO UPDATE SET
			   delta = $3`,
				metric.ID,
				metric.MType,
				*metric.Delta,
				nil,
				metric.Hash(),
			)

			return err
		})
	case model.Gauge:
		return ps.retryableHelper.Run(ctx, func() error {
			_, err := ps.db.ExecContext(
				ctx,
				`INSERT INTO metric (id, mtype, delta, value, hash) VALUES ($1, $2, $3, $4, $5) 
			   ON CONFLICT (mtype, hash) DO UPDATE SET
			   value = $4`,
				metric.ID,
				metric.MType,
				nil,
				*metric.Value,
				metric.Hash(),
			)

			return err
		})
	}

	return nil
}

func (ps *PostgresStorage) UpdateAndSaveBatch(ctx context.Context, metrics []*model.Metric) error {
	var (
		counters, gauges []*model.Metric
	)

	for _, metric := range metrics {
		switch metric.MType {
		case model.Counter:
			counters = append(counters, metric)
		case model.Gauge:
			gauges = append(gauges, metric)
		}
	}

	insertData := make(map[string]map[string]interface{}, len(counters)+len(gauges))

	for _, counter := range counters {
		delta := *counter.Delta

		if val, ok := insertData["c"+counter.Hash()]; ok {
			delta = *(val["delta"].(*int64)) + delta
		}

		insertData["c"+counter.Hash()] = map[string]interface{}{
			"id":    counter.ID,
			"mtype": counter.MType,
			"delta": &delta,
			"value": nil,
			"hash":  counter.Hash(),
		}
	}

	for _, gauge := range gauges {
		insertData["g"+gauge.Hash()] = map[string]interface{}{
			"id":    gauge.ID,
			"mtype": gauge.MType,
			"delta": nil,
			"value": *gauge.Value,
			"hash":  gauge.Hash(),
		}
	}

	dbx := sqlx.NewDb(ps.db, "postgres")

	return ps.retryableHelper.Run(ctx, func() error {
		tx, err := dbx.BeginTxx(ctx, nil)
		if err != nil {
			return err
		}

		_, err = tx.NamedExec(
			`INSERT INTO metric (id, mtype, delta, value, hash) VALUES (:id, :mtype, :delta, :value, :hash) 
			   ON CONFLICT (mtype, hash) DO UPDATE SET
			   delta = metric.delta + EXCLUDED.delta, value = EXCLUDED.value`,
			slices.Collect(maps.Values(insertData)),
		)

		if err != nil {
			errRollback := tx.Rollback()
			if errRollback != nil {

				return fmt.Errorf("rollback failed: %w", errors.Join(err, errRollback))
			}

			return err
		}

		return tx.Commit()
	})
}

func NewPostgresStorage(db *sql.DB, logger *zerolog.Logger) (*PostgresStorage, error) {
	err := migrations.RunMigrations(db)
	if err != nil {
		return nil, err
	}

	retryableHelper := helper.NewRetryableHelper(
		&helper.StandardSleeper{},
		func(err error) bool {
			var pgErr *pgconn.PgError

			if errors.As(err, &pgErr) {
				switch pgErr.Code {
				case pgerrcode.ConnectionException,
					pgerrcode.ConnectionDoesNotExist,
					pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection,
					pgerrcode.ConnectionFailure,
					pgerrcode.SQLClientUnableToEstablishSQLConnection,
					pgerrcode.TransactionResolutionUnknown,
					pgerrcode.ProtocolViolation:
					return true
				}
			}

			return false
		},
		[]time.Duration{
			time.Second,
			time.Second * 3,
			time.Second * 5,
		},
		logger,
	)

	return &PostgresStorage{
		db:              db,
		logger:          logger,
		retryableHelper: retryableHelper,
	}, nil
}

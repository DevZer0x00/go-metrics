package repository

import (
	"context"
	"database/sql"
	"errors"
	"go-metrics/internal/model"
	"maps"
	"slices"

	"github.com/jmoiron/sqlx"
)

type PostgresStorage struct {
	db  *sql.DB
	ctx context.Context
}

func (ps *PostgresStorage) GetOrRegister(mtype, name string) (*model.Metric, error) {
	row := ps.db.QueryRowContext(
		ps.ctx,
		"SELECT delta, value FROM metric WHERE hash = $1 AND mtype = $2 ",
		model.GetMetricHash(name),
		mtype,
	)

	var (
		delta sql.NullInt64
		value sql.NullFloat64
	)

	err := row.Scan(&delta, &value)

	if errors.Is(err, sql.ErrNoRows) {
		return model.NewMetric(name, mtype), nil
	}

	metric := model.NewMetric(name, mtype)

	switch mtype {
	case model.Counter:
		metric.Delta = &delta.Int64
	case model.Gauge:
		metric.Value = &value.Float64
	}

	return metric, err
}

func (ps *PostgresStorage) All() ([]*model.Metric, error) {
	rows, err := ps.db.QueryContext(
		ps.ctx,
		"SELECT id, mtype, delta, value FROM metric ORDER BY mtype, hash",
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	models := make([]*model.Metric, 0)

	for rows.Next() {
		metric := &model.Metric{}
		err := rows.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
		if err != nil {
			return nil, err
		}

		models = append(models, metric)
	}

	return models, nil
}

func (ps *PostgresStorage) Has(mtype, name string) (bool, error) {
	row := ps.db.QueryRowContext(
		ps.ctx,
		"SELECT 1 FROM metric WHERE hash = $1 AND mtype = $2 ",
		model.GetMetricHash(name),
		mtype,
	)

	if errors.Is(row.Scan(), sql.ErrNoRows) {
		return false, nil
	}

	return true, nil
}

func (ps *PostgresStorage) Save(metric *model.Metric) error {
	var err error

	switch metric.MType {
	case model.Counter:
		_, err = ps.db.Exec(
			`INSERT INTO metric (id, mtype, delta, value, hash) VALUES ($1, $2, $3, $4, $5) 
			   ON CONFLICT (mtype, hash) DO UPDATE SET
			   delta = $3`,
			metric.ID,
			metric.MType,
			*metric.Delta,
			nil,
			metric.Hash(),
		)
	case model.Gauge:
		_, err = ps.db.Exec(
			`INSERT INTO metric (id, mtype, delta, value, hash) VALUES ($1, $2, $3, $4, $5) 
			   ON CONFLICT (mtype, hash) DO UPDATE SET
			   delta = $3`,
			metric.ID,
			metric.MType,
			nil,
			*metric.Value,
			metric.Hash(),
		)
	}

	return err
}

func (ps *PostgresStorage) SaveBatch(metrics []*model.Metric) error {
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

	dbx := sqlx.NewDb(ps.db, "postgres")
	tx, err := dbx.Beginx()
	if err != nil {
		return err
	}

	if len(counters) > 0 {
		insertData := make(map[string]map[string]interface{}, len(counters))

		for _, counter := range counters {
			delta := *counter.Delta

			if val, ok := insertData[counter.Hash()]; ok {
				delta = *(val["delta"].(*int64)) + delta
			}

			insertData[counter.Hash()] = map[string]interface{}{
				"id":    counter.ID,
				"mtype": counter.MType,
				"delta": &delta,
				"hash":  counter.Hash(),
			}
		}

		_, err := tx.NamedExec(
			`INSERT INTO metric (id, mtype, delta, hash) VALUES (:id, :mtype, :delta, :hash) 
			   ON CONFLICT (mtype, hash) DO UPDATE SET
			   delta = metric.delta + EXCLUDED.delta`,
			slices.Collect(maps.Values(insertData)),
		)
		if err != nil {
			return err
		}
	}

	if len(gauges) > 0 {
		insertData := make(map[string]map[string]interface{}, len(gauges))

		for _, gauge := range gauges {
			insertData[gauge.Hash()] = map[string]interface{}{
				"id":    gauge.ID,
				"mtype": gauge.MType,
				"value": *gauge.Value,
				"hash":  gauge.Hash(),
			}
		}

		_, err := tx.NamedExec(
			`INSERT INTO metric (id, mtype, value, hash) VALUES (:id, :mtype, :value, :hash) 
			   ON CONFLICT (mtype, hash) DO UPDATE SET
			   value = EXCLUDED.value`,
			slices.Collect(maps.Values(insertData)),
		)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func NewPostgresStorage(ctx context.Context, db *sql.DB) *PostgresStorage {
	return &PostgresStorage{
		db:  db,
		ctx: ctx,
	}
}

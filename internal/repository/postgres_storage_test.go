package repository

import (
	"database/sql"
	"go-metrics/internal/model"
	"go-metrics/migrations"
	"os"
	"strconv"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jaswdr/faker/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func openAndPrepareDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL environment variable not set")
	}

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)

	require.NoError(t, db.PingContext(t.Context()))

	require.NoError(t, migrations.RunMigrations(t.Context(), db))

	_, err = db.Exec("TRUNCATE TABLE metric")
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	return db
}

func TestPostgresStorageAll(t *testing.T) {
	fake := faker.New()

	db := openAndPrepareDB(t)
	storage := NewPostgresStorage(t.Context(), db)

	all, err := storage.All()
	require.NoError(t, err)
	assert.Len(t, all, 0)

	for i := range 5 {
		id := fake.Lorem().Word() + strconv.FormatInt(int64(i), 10)
		hash := model.GetMetricHash(id)

		_, err := db.Exec(
			"INSERT INTO metric (id, mtype, delta, value, hash) VALUES ($1, $2, $3, $4, $5)",
			id,
			model.Counter,
			i,
			nil,
			hash,
		)
		require.NoError(t, err)
	}

	metrics, err := storage.All()
	require.NoError(t, err)

	assert.Len(t, metrics, 5)
}

func TestPostgresStorageSave(t *testing.T) {
	db := openAndPrepareDB(t)
	storage := NewPostgresStorage(t.Context(), db)
	metric := model.NewMetric("test", model.Counter)
	metric.UpdateDelta(10)
	require.NoError(t, storage.Save(metric))

	metric.UpdateDelta(20)
	require.NoError(t, storage.Save(metric))

	metrics, err := storage.All()
	require.NoError(t, err)
	assert.Len(t, metrics, 1)

	metricDB := metrics[0]
	assert.Equal(t, metric.ID, metricDB.ID)
	assert.Equal(t, metric.MType, metricDB.MType)
	assert.Equal(t, metric.Delta, metricDB.Delta)
	assert.Equal(t, metric.Value, metricDB.Value)
	assert.Equal(t, metric.HashValue, metricDB.HashValue)

	metric = model.NewMetric("test", model.Gauge)
	metric.UpdateValue(10.12)
	require.NoError(t, storage.Save(metric))

	metrics, err = storage.All()
	require.NoError(t, err)
	assert.Len(t, metrics, 2)

	metricDB = metrics[1]
	assert.Equal(t, metric.ID, metricDB.ID)
	assert.Equal(t, metric.MType, metricDB.MType)
	assert.Equal(t, metric.Delta, metricDB.Delta)
	assert.Equal(t, metric.Value, metricDB.Value)
	assert.Equal(t, metric.HashValue, metricDB.HashValue)
}

func TestPostgresStorageHas(t *testing.T) {
	db := openAndPrepareDB(t)
	storage := NewPostgresStorage(t.Context(), db)

	has, err := storage.Has(model.Gauge, "test")
	require.NoError(t, err)
	assert.False(t, has)

	metric := model.NewMetric("test", model.Gauge)
	metric.UpdateValue(10.12)
	require.NoError(t, storage.Save(metric))

	has, err = storage.Has(model.Gauge, "test")
	require.NoError(t, err)
	assert.True(t, has)
}

func TestPostgresStorageGetOrRegister(t *testing.T) {
	db := openAndPrepareDB(t)
	storage := NewPostgresStorage(t.Context(), db)

	id := "test"
	metric, err := storage.GetOrRegister(model.Counter, id)
	require.NoError(t, err)

	assert.Equal(t, "test", metric.ID)
	assert.Equal(t, int64(0), *metric.Delta)

	hash := model.GetMetricHash(id)

	_, err = db.Exec(
		"INSERT INTO metric (id, mtype, delta, value, hash) VALUES ($1, $2, $3, $4, $5)",
		id,
		model.Counter,
		10,
		nil,
		hash,
	)
	require.NoError(t, err)

	metric, err = storage.GetOrRegister(model.Counter, id)
	require.NoError(t, err)

	assert.Equal(t, int64(10), *metric.Delta)
}

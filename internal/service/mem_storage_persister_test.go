package service

import (
	"fmt"
	"go-metrics/internal/model"
	"go-metrics/internal/repository"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetricsPersister(t *testing.T) {
	storage := &repository.MemStorage{}
	logger := zerolog.New(io.Discard)
	persister := NewMetricsPersister(storage, &logger, true, "1.db")
	assert.Equal(t, storage, persister.storage)
	assert.True(t, persister.loadOnInit)
	assert.Equal(t, "1.db", persister.storageFilePath)
}

func TestMemMetricsPersisterInitWithoutLoad(t *testing.T) {
	storage := repository.NewMemStorage()

	logger := zerolog.New(io.Discard)
	persister := NewMetricsPersister(
		storage,
		&logger,
		false,
		filepath.Join(
			os.TempDir(),
			fmt.Sprintf("service%s.db", time.Now().Format("20060102150405")),
		),
	)
	err := persister.Init(t.Context())
	require.NoError(t, err)

	stat, err := os.Stat(persister.storageFilePath)
	require.NoError(t, err)
	assert.Equal(t, int64(0), stat.Size())
}

func TestMemMetricsPersisterInitWithLoad(t *testing.T) {
	dbFile, err := os.CreateTemp("", "server*.db")
	require.NoError(t, err)

	defer func(filename string) {
		_ = os.Remove(filename)
	}(dbFile.Name())

	_, err = dbFile.Write([]byte(`
[
  {"id":"LastGC","type":"gauge","value":1257894000000000000},
  {"id":"NumGC","type":"counter","delta":42}
]
`))
	require.NoError(t, err)
	err = dbFile.Close()
	require.NoError(t, err)

	storage := repository.NewMemStorage()
	logger := zerolog.New(io.Discard)
	persister := NewMetricsPersister(
		storage,
		&logger,
		true,
		dbFile.Name(),
	)

	err = persister.Init(t.Context())
	require.NoError(t, err)

	metrics, err := storage.All(t.Context())
	require.NoError(t, err)
	assert.Len(t, metrics, 2)

	has, _ := storage.Has(t.Context(), model.Gauge, "LastGC")
	assert.True(t, has)
}

func TestMemMetricsPersisterFlush(t *testing.T) {
	storage := repository.NewMemStorage()
	logger := zerolog.New(io.Discard)
	persister := NewMetricsPersister(
		storage,
		&logger,
		true,
		filepath.Join(
			os.TempDir(),
			fmt.Sprintf("service%s.db", time.Now().Format("20060102150405")),
		),
	)

	err := persister.Init(t.Context())
	require.NoError(t, err)

	err = storage.Save(t.Context(), model.NewMetric("LastGC", "counter"))
	require.NoError(t, err)
	err = storage.Save(t.Context(), model.NewMetric("test", "gauge"))
	require.NoError(t, err)

	err = persister.Flush(t.Context())
	require.NoError(t, err)

	err = persister.Init(t.Context())
	require.NoError(t, err)
}

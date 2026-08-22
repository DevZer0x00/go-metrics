package repository

import (
	"go-metrics/internal/model"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOrRegisterCounter(t *testing.T) {
	storage := NewMemStorage()
	require.NotNil(t, storage.counter)

	metrics, err := storage.GetOrRegister(model.Counter, "test")
	require.NoError(t, err)
	metrics, err = storage.GetOrRegister(model.Counter, "test")
	require.NoError(t, err)

	assert.Len(t, storage.counter, 1)
	assert.Len(t, storage.gauge, 0)

	require.NotNil(t, metrics)
	assert.NotNil(t, metrics.Delta)

	assert.Contains(t, storage.counter, metrics.Hash)
}

func TestGetOrRegisterGauge(t *testing.T) {
	storage := NewMemStorage()
	require.NotNil(t, storage.gauge)

	metrics, err := storage.GetOrRegister(model.Gauge, "test")
	require.NoError(t, err)
	metrics, err = storage.GetOrRegister(model.Gauge, "test")
	require.NoError(t, err)

	assert.Len(t, storage.gauge, 1)
	assert.Len(t, storage.counter, 0)

	require.NotNil(t, metrics)
	assert.NotNil(t, metrics.Value)

	assert.Contains(t, storage.gauge, metrics.Hash)
}

func TestAll(t *testing.T) {
	storage := NewMemStorage()

	_, err := storage.GetOrRegister(model.Counter, "test1")
	require.NoError(t, err)
	_, err = storage.GetOrRegister(model.Counter, "test2")
	require.NoError(t, err)
	_, err = storage.GetOrRegister(model.Gauge, "test1")
	require.NoError(t, err)

	all, err := storage.All()
	require.NoError(t, err)

	assert.Len(t, all, 3)
}

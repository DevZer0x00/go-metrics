package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCounterMetrics(t *testing.T) {
	metrics := NewMetric("test", Counter)

	assert.NotNil(t, metrics)
	assert.Equal(t, "test", metrics.ID)
	assert.Equal(t, Counter, metrics.MType)
	assert.Equal(t, int64(0), *metrics.Delta)
	assert.NotEmpty(t, metrics.Hash)
}

func TestNewGaugeMetrics(t *testing.T) {
	metrics := NewMetric("test", Gauge)

	assert.NotNil(t, metrics)
	assert.Equal(t, "test", metrics.ID)
	assert.Equal(t, Gauge, metrics.MType)
	assert.Equal(t, float64(0), *metrics.Value)
	assert.NotEmpty(t, metrics.Hash)
}

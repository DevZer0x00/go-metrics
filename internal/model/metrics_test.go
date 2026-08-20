package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCounterMetrics(t *testing.T) {
	metrics := NewMetrics("test", Counter)

	assert.NotNil(t, metrics)
	assert.Equal(t, "test", metrics.ID)
	assert.Equal(t, Counter, metrics.MType)
	assert.Equal(t, int64(0), *metrics.Delta)
	assert.NotEmpty(t, metrics.Hash)
}

func TestNewGaugeMetrics(t *testing.T) {
	metrics := NewMetrics("test", Counter)

	assert.NotNil(t, metrics)
	assert.Equal(t, "test", metrics.ID)
	assert.Equal(t, Counter, metrics.MType)
	assert.Equal(t, int64(0), *metrics.Delta)
	assert.NotEmpty(t, metrics.Hash)
}

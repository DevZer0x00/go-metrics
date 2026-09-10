package service

import (
	"go-metrics/internal/model"
	"io"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestingRepository struct {
	getOrRegisterFunc func(string, string) (*model.Metric, error)
	hasFunc           func(string, string) (bool, error)
	allFunc           func() ([]*model.Metric, error)
	saveFunc          func(*model.Metric) error
}

func (r *TestingRepository) GetOrRegister(mtype, name string) (*model.Metric, error) {
	return r.getOrRegisterFunc(mtype, name)
}

func (r *TestingRepository) Has(mtype, name string) (bool, error) {
	return r.hasFunc(mtype, name)
}

func (r *TestingRepository) All() ([]*model.Metric, error) {
	return r.allFunc()
}

func (r *TestingRepository) Save(metric *model.Metric) error {
	return r.saveFunc(metric)
}

type TestingPersister struct {
	InitFunc  func() error
	FlushFunc func() error
}

func (p *TestingPersister) Init() error {
	return p.InitFunc()
}

func (p *TestingPersister) Flush() error {
	return p.FlushFunc()
}

func TestMetricsServiceGet(t *testing.T) {
	rep := &TestingRepository{
		getOrRegisterFunc: func(mtype, name string) (*model.Metric, error) {
			return &model.Metric{
				ID:    name,
				MType: mtype,
			}, nil
		},
		hasFunc: func(mtype, name string) (bool, error) {
			if mtype == model.Counter {
				return true, nil
			}

			return false, nil
		},
	}

	logger := zerolog.New(io.Discard)
	service := NewMetricsService(rep, &TestingPersister{}, &logger)
	metric, err := service.Get(model.Counter, "counter")
	assert.NoError(t, err)
	assert.NotNil(t, metric)

	_, err = service.Get(model.Gauge, "gauge")
	assert.ErrorAs(t, err, &ErrMetricNotFound)
}

func TestMetricsServiceGetAll(t *testing.T) {
	rep := &TestingRepository{
		allFunc: func() ([]*model.Metric, error) {
			return make([]*model.Metric, 0), nil
		},
	}

	logger := zerolog.New(io.Discard)
	service := NewMetricsService(rep, &TestingPersister{}, &logger)
	all, err := service.GetAll()
	assert.NoError(t, err)
	assert.Len(t, all, 0)
}

func TestUpdateFromStringValueCounter(t *testing.T) {
	rep := &TestingRepository{
		getOrRegisterFunc: func(mtype, name string) (*model.Metric, error) {
			return &model.Metric{
				ID:    name,
				MType: mtype,
				Delta: new(int64),
			}, nil
		},
		saveFunc: func(m *model.Metric) error {
			assert.Equal(t, "test", m.ID)
			assert.Equal(t, model.Counter, m.MType)
			require.NotNil(t, m.Delta)
			assert.Equal(t, int64(10), *m.Delta)

			return nil
		},
	}

	logger := zerolog.New(io.Discard)
	service := NewMetricsService(rep, &TestingPersister{}, &logger)
	err := service.UpdateFromStringValue(model.Counter, "test", "10")
	require.NoError(t, err)
}

func TestUpdateFromStringValueGauge(t *testing.T) {
	rep := &TestingRepository{
		getOrRegisterFunc: func(mtype, name string) (*model.Metric, error) {
			return &model.Metric{
				ID:    name,
				MType: mtype,
				Value: new(float64),
			}, nil
		},
		saveFunc: func(m *model.Metric) error {
			assert.Equal(t, "test", m.ID)
			assert.Equal(t, model.Gauge, m.MType)
			require.NotNil(t, m.Value)
			assert.Equal(t, 10.12, *m.Value)

			return nil
		},
	}

	logger := zerolog.New(io.Discard)
	service := NewMetricsService(rep, &TestingPersister{}, &logger)
	err := service.UpdateFromStringValue(model.Gauge, "test", "10.12")
	require.NoError(t, err)
}

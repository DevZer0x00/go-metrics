package repository

import (
	"go-metrics/internal/model"
)

type MemStorage struct {
	counter map[string]*model.Metrics
	gauge   map[string]*model.Metrics
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		counter: make(map[string]*model.Metrics),
		gauge:   make(map[string]*model.Metrics),
	}
}

func (ms *MemStorage) GetOrRegisterCounter(name string) (*model.Metrics, error) {
	hash := model.GetMetricsHash(name)

	metrics, ok := ms.counter[hash]
	if !ok {
		metrics = model.NewMetrics(name, model.Counter)
		err := ms.Save(metrics)
		if err != nil {
			return nil, err
		}
	}

	return metrics, nil
}

func (ms *MemStorage) GetOrRegisterGauge(name string) (*model.Metrics, error) {
	hash := model.GetMetricsHash(name)

	metrics, ok := ms.gauge[hash]
	if !ok {
		metrics = model.NewMetrics(name, model.Gauge)
		err := ms.Save(metrics)
		if err != nil {
			return nil, err
		}
	}

	return metrics, nil
}

func (ms *MemStorage) Save(metrics *model.Metrics) error {
	switch metrics.MType {
	case model.Gauge:
		ms.gauge[metrics.Hash] = metrics
	case model.Counter:
		ms.counter[metrics.Hash] = metrics
	}

	return nil
}

package repository

import (
	"go-metrics/internal/model"
	"maps"
	"slices"
)

type MemStorage struct {
	counter map[string]*model.Metric
	gauge   map[string]*model.Metric
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		counter: make(map[string]*model.Metric),
		gauge:   make(map[string]*model.Metric),
	}
}

func (ms *MemStorage) GetOrRegisterCounter(name string) (*model.Metric, error) {
	hash := model.GetMetricHash(name)

	metrics, ok := ms.counter[hash]
	if !ok {
		metrics = model.NewMetric(name, model.Counter)
		err := ms.Save(metrics)
		if err != nil {
			return nil, err
		}
	}

	return metrics, nil
}

func (ms *MemStorage) GetOrRegisterGauge(name string) (*model.Metric, error) {
	hash := model.GetMetricHash(name)

	metrics, ok := ms.gauge[hash]
	if !ok {
		metrics = model.NewMetric(name, model.Gauge)
		err := ms.Save(metrics)
		if err != nil {
			return nil, err
		}
	}

	return metrics, nil
}

func (ms *MemStorage) Save(metrics *model.Metric) error {
	switch metrics.MType {
	case model.Gauge:
		ms.gauge[metrics.Hash] = metrics
	case model.Counter:
		ms.counter[metrics.Hash] = metrics
	}

	return nil
}

func (ms *MemStorage) HasCounter(name string) (bool, error) {
	hash := model.GetMetricHash(name)
	_, ok := ms.counter[hash]

	return ok, nil
}

func (ms *MemStorage) HasGauge(name string) (bool, error) {
	hash := model.GetMetricHash(name)
	_, ok := ms.gauge[hash]

	return ok, nil
}

func (ms *MemStorage) All() ([]*model.Metric, error) {
	res := make([]*model.Metric, 0, len(ms.counter)+len(ms.gauge))

	res = append(res, slices.Collect(maps.Values(ms.counter))...)
	res = append(res, slices.Collect(maps.Values(ms.gauge))...)

	return res, nil
}

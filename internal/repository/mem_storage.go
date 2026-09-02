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

func (ms *MemStorage) getMetricMapByType(mtype string) map[string]*model.Metric {
	var metricMap map[string]*model.Metric

	switch mtype {
	case model.Counter:
		metricMap = ms.counter
	case model.Gauge:
		metricMap = ms.gauge
	}

	return metricMap
}

func (ms *MemStorage) GetOrRegister(mtype, name string) (*model.Metric, error) {
	hash := model.GetMetricHash(name)

	var err error

	metrics, ok := ms.getMetricMapByType(mtype)[hash]
	if !ok {
		metrics = model.NewMetric(name, mtype)
		err = ms.Save(metrics)
		if err != nil {
			return nil, err
		}
	}

	return metrics, nil
}

func (ms *MemStorage) Save(metrics *model.Metric) error {
	ms.getMetricMapByType(metrics.MType)[metrics.Hash] = metrics

	return nil
}

func (ms *MemStorage) Has(mtype, name string) (bool, error) {
	hash := model.GetMetricHash(name)
	_, ok := ms.getMetricMapByType(mtype)[hash]

	return ok, nil
}

func (ms *MemStorage) All() ([]*model.Metric, error) {
	res := make([]*model.Metric, 0, len(ms.counter)+len(ms.gauge))

	res = append(res, slices.Collect(maps.Values(ms.counter))...)
	res = append(res, slices.Collect(maps.Values(ms.gauge))...)

	return res, nil
}

package repository

import "go-metrics/internal/model"

type MetricRepository interface {
	GetOrRegisterCounter(name string) (*model.Metric, error)
	GetOrRegisterGauge(name string) (*model.Metric, error)
	HasCounter(name string) (bool, error)
	HasGauge(name string) (bool, error)
	All() ([]*model.Metric, error)
	Save(metrics *model.Metric) error
}

var Metric MetricRepository = NewMemStorage()

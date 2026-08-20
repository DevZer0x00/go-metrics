package repository

import "go-metrics/internal/model"

type MetricsRepository interface {
	GetOrRegisterCounter(name string) (*model.Metrics, error)
	GetOrRegisterGauge(name string) (*model.Metrics, error)
	Save(metrics *model.Metrics) error
}

var Metrics = NewMemStorage()

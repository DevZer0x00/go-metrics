package repository

import "go-metrics/internal/model"

type MetricRepository interface {
	GetOrRegister(mtype, name string) (*model.Metric, error)
	Has(mtype, name string) (bool, error)
	All() ([]*model.Metric, error)
	Save(metric *model.Metric) error
}

var Metric MetricRepository = NewMemStorage()

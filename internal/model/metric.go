package model

import (
	"crypto/md5"
	"fmt"
)

const (
	Counter = "counter"
	Gauge   = "gauge"
)

type Metric struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

func (metrics *Metric) UpdateCounter(value int64) int64 {
	*metrics.Delta += value

	return *metrics.Delta
}

func (metrics *Metric) UpdateGauge(value float64) float64 {
	*metrics.Value = value

	return *metrics.Value
}

func NewMetric(name string, mtype string) *Metric {
	metrics := &Metric{
		ID:    name,
		MType: mtype,
		Hash:  GetMetricHash(name),
	}

	switch mtype {
	case Counter:
		metrics.Delta = new(int64)
	case Gauge:
		metrics.Value = new(float64)
	default:
		panic(fmt.Sprintf("unknown metrics type: %s", mtype))
	}

	return metrics
}

func GetMetricHash(name string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(name)))
}

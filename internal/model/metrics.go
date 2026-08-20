package model

import (
	"crypto/md5"
	"fmt"
)

const (
	Counter = "counter"
	Gauge   = "gauge"
)

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

func (metrics *Metrics) UpdateCounter(value int64) int64 {
	*metrics.Delta += value

	return *metrics.Delta
}

func (metrics *Metrics) UpdateGauge(value float64) float64 {
	*metrics.Value = value

	return *metrics.Value
}

func NewMetrics(name string, mtype string) *Metrics {
	metrics := &Metrics{
		ID:    name,
		MType: mtype,
		Hash:  GetMetricsHash(name),
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

func GetMetricsHash(name string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(name)))
}

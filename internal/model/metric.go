package model

import (
	"crypto/md5"
	"errors"
	"fmt"
	"strconv"
	"strings"
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

func (metrics *Metric) UpdateDelta(value int64) int64 {
	*metrics.Delta += value

	return *metrics.Delta
}

func (metrics *Metric) UpdateValue(value float64) float64 {
	*metrics.Value = value

	return *metrics.Value
}

func (metric *Metric) ValueToString() string {
	switch metric.MType {
	case Counter:
		return strconv.FormatInt(*metric.Delta, 10)
	case Gauge:
		return strings.TrimSuffix(strconv.FormatFloat(*metric.Value, 'f', 3, 64), "0")
	}

	return ""
}

func NewMetric(name string, mtype string) (*Metric, error) {
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
		return nil, errors.New(fmt.Sprintf("unknown metrics type: %s", mtype))
	}

	return metrics, nil
}

func GetMetricHash(name string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(name)))
}

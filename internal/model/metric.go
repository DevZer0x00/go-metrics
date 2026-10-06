package model

import (
	"crypto/md5"
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
	hash  string
}

func (metric *Metric) UpdateDelta(value int64) int64 {
	*metric.Delta += value

	return *metric.Delta
}

func (metric *Metric) UpdateValue(value float64) float64 {
	*metric.Value = value

	return *metric.Value
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

func (metric *Metric) Hash() string {
	if len(metric.hash) == 0 {
		metric.hash = GetMetricHash(metric.ID)
	}

	return metric.hash
}

func NewMetric(name string, mtype string) *Metric {
	metric := &Metric{
		ID:    name,
		MType: mtype,
	}
	metric.hash = GetMetricHash(metric.ID)

	switch mtype {
	case Counter:
		metric.Delta = new(int64)
	case Gauge:
		metric.Value = new(float64)
	}

	return metric
}

func GetMetricHash(name string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(name)))
}

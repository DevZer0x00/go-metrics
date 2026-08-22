package service

import (
	"errors"
	"fmt"
	"go-metrics/internal/model"
	"strconv"
)

type MetricRepository interface {
	GetOrRegister(mtype, name string) (*model.Metric, error)
	Has(mtype, name string) (bool, error)
	All() ([]*model.Metric, error)
	Save(metric *model.Metric) error
}

var ErrInvalidMetricValue = errors.New("invalid metric value")
var ErrInvalidMetricType = errors.New("invalid metric type")
var ErrMetricNotFound = errors.New("metric not found")

var allowedMetricTypes = map[string]bool{
	model.Counter: true,
	model.Gauge:   true,
}

type MetricsService struct {
	repository MetricRepository
}

func CheckMetricType(metricType string) error {
	if _, exists := allowedMetricTypes[metricType]; !exists {
		return ErrInvalidMetricType
	}

	return nil
}

func (service *MetricsService) Get(metricType, metricName string) (*model.Metric, error) {
	if has, err := service.repository.Has(metricType, metricName); err != nil {
		return nil, err
	} else if !has {
		return nil, ErrMetricNotFound
	}

	return service.repository.GetOrRegister(metricType, metricName)
}

func (service *MetricsService) GetAll() ([]*model.Metric, error) {
	return service.repository.All()
}

func (service *MetricsService) UpdateFromStringValue(metricType, metricName, metricValue string) error {
	err := CheckMetricType(metricType)
	if err != nil {
		return err
	}

	metrics, err := service.repository.GetOrRegister(metricType, metricName)
	if err != nil {
		return fmt.Errorf("error getting metrics from repository: %w", err)
	}

	switch metricType {
	case model.Counter:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return ErrInvalidMetricValue
		}

		metrics.UpdateDelta(value)
	case model.Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return ErrInvalidMetricValue
		}

		metrics.UpdateValue(value)
	}

	err = service.repository.Save(metrics)
	if err != nil {
		return fmt.Errorf("error save metrics: %w", err)
	}

	return nil
}

func NewMetricsService(repository MetricRepository) *MetricsService {
	return &MetricsService{
		repository: repository,
	}
}

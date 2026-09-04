package service

import (
	"errors"
	"fmt"
	"go-metrics/internal/model"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/non-standard/validators"
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

var validate *validator.Validate

func init() {
	validate = validator.New()
	_ = validate.RegisterValidation("notblank", validators.NotBlank)
	_ = validate.RegisterValidation(
		"checkMetricValue",
		func(fl validator.FieldLevel) bool {
			parentField := fl.Parent()

			switch parentField.FieldByName("MType").String() {
			case model.Counter:
				return !parentField.FieldByName("Delta").IsNil()
			case model.Gauge:
				return !parentField.FieldByName("Value").IsNil()
			}

			return true
		},
		true,
	)

	validate.RegisterStructValidationMapRules(
		map[string]string{
			"ID":    "required,notblank",
			"MType": fmt.Sprintf("required,oneof=%s %s", model.Counter, model.Gauge),
			"Delta": "checkMetricValue",
			"Value": "checkMetricValue",
		},
		model.Metric{},
	)
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
	metric := model.NewMetric(metricName, metricType)

	switch metricType {
	case model.Counter:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return ErrInvalidMetricValue
		}

		*metric.Delta = value
	case model.Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return ErrInvalidMetricValue
		}

		*metric.Value = value
	}

	return service.UpdateFromModel(metric)
}

func (service *MetricsService) UpdateFromModel(metricReq *model.Metric) error {
	err := validate.Struct(metricReq)
	if err != nil {
		return err
	}

	metric, err := service.repository.GetOrRegister(metricReq.MType, metricReq.ID)
	if err != nil {
		return fmt.Errorf("error getting metrics from repository: %w", err)
	}

	switch metricReq.MType {
	case model.Counter:
		metric.UpdateDelta(*metricReq.Delta)
	case model.Gauge:
		metric.UpdateValue(*metricReq.Value)
	}

	err = service.repository.Save(metric)
	if err != nil {
		return fmt.Errorf("error save metric: %w", err)
	}

	return nil
}

func NewMetricsService(repository MetricRepository) *MetricsService {
	return &MetricsService{
		repository: repository,
	}
}

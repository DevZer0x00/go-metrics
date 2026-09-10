package service

import (
	"errors"
	"fmt"
	"go-metrics/internal/model"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/non-standard/validators"
	"github.com/rs/zerolog"
)

type MetricRepository interface {
	GetOrRegister(mtype, name string) (*model.Metric, error)
	Has(mtype, name string) (bool, error)
	All() ([]*model.Metric, error)
	Save(metric *model.Metric) error
}

type MetricPersister interface {
	Init() error
	Flush() error
}

var ErrInvalidMetricValue = errors.New("invalid metric value")
var ErrInvalidMetricType = errors.New("invalid metric type")
var ErrMetricNotFound = errors.New("metric not found")

var allowedMetricTypes = map[string]bool{
	model.Counter: true,
	model.Gauge:   true,
}

type MetricsService struct {
	repository    MetricRepository
	persister     MetricPersister
	validate      *validator.Validate
	logger        *zerolog.Logger
	flushTicker   *time.Ticker
	syncPersister bool
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
	err := service.validate.Struct(metricReq)
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

	if service.syncPersister {
		service.flushPersister()
	}

	return nil
}

func (service *MetricsService) flushPersister() {
	err := service.persister.Flush()
	if err != nil {
		service.logger.Error().Err(err).Msg("failed to flush memory storage")
	}
}

func (service *MetricsService) InitPersister(persistInterval uint64) error {
	err := service.persister.Init()
	if err != nil {
		return err
	}

	if persistInterval > 0 {
		service.flushTicker = time.NewTicker(time.Second * time.Duration(persistInterval))

		go func() {
			for range service.flushTicker.C {
				service.flushPersister()
			}
		}()
	} else {
		service.syncPersister = true
	}

	return nil
}

func (service *MetricsService) Close() {
	if service.flushTicker != nil {
		service.flushTicker.Stop()
	}

	service.flushPersister()
}

func NewMetricsService(repository MetricRepository, persister MetricPersister, logger *zerolog.Logger) *MetricsService {
	validate := validator.New()
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

	return &MetricsService{
		repository: repository,
		persister:  persister,
		validate:   validate,
		logger:     logger,
	}
}

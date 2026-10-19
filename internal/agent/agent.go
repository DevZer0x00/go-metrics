package agent

import (
	"context"
	"errors"
	"fmt"
	"go-metrics/internal/config"
	"go-metrics/internal/helper"
	"go-metrics/internal/model"
	"go-metrics/internal/resty/middleware"
	"math/rand"
	"net"
	"time"

	"github.com/rs/zerolog"
	"resty.dev/v3"
)

type MetricsAgent struct {
	httpClient      *resty.Client
	serverAddr      *config.ServerAddr
	pollCount       int64
	randomValue     float64
	metrics         []MemMetrics
	retryableHelper *helper.RetryableHelper
	logger          *zerolog.Logger
}

func (a *MetricsAgent) resetAfterSend() {
	a.pollCount = 0
	a.randomValue = 0
	a.metrics = make([]MemMetrics, 0)
}

func (a *MetricsAgent) Collect() error {
	a.pollCount++
	a.randomValue = rand.Float64() * 100
	metrics, err := CollectMemMetrics()
	if err != nil {
		return err
	}

	a.metrics = metrics

	return nil
}

func (a *MetricsAgent) sendReports(ctx context.Context, metrics []*model.Metric) error {
	return a.retryableHelper.Run(ctx, func() error {
		_, err := a.httpClient.R().
			SetContext(ctx).
			SetBody(metrics).
			Post(fmt.Sprintf("http://%s/updates/", a.serverAddr.Addr))
		if err != nil {
			return fmt.Errorf("error making request: %w", err)
		}

		return nil
	})
}

func (a *MetricsAgent) SendBatch(ctx context.Context) {
	if a.pollCount == 0 {
		return
	}

	metrics := make([]*model.Metric, len(a.metrics)+2)

	key := 0

	for _, memMetric := range a.metrics {
		metric := &model.Metric{
			ID:    memMetric.Label,
			MType: model.Gauge,
			Value: &memMetric.Value,
		}

		metrics[key] = metric
		key++
	}

	delta := a.pollCount
	metrics[key] = &model.Metric{
		ID:    "PollCount",
		MType: model.Counter,
		Delta: &delta,
	}
	key++

	randomValue := a.randomValue
	metrics[key] = &model.Metric{
		ID:    "RandomValue",
		MType: model.Gauge,
		Value: &randomValue,
	}

	err := a.sendReports(ctx, metrics)
	if err != nil {
		a.logger.
			Err(err).
			Msg("error sending metrics")
	}

	a.resetAfterSend()
}

func NewMetricsAgent(client *resty.Client, addr *config.ServerAddr, logger *zerolog.Logger) *MetricsAgent {
	client.AddContentTypeEncoder("application/json", resty.InMemoryJSONMarshal)
	client.AddRequestMiddleware(middleware.RequestGzipCompress(logger))

	retryableHelper := helper.NewRetryableHelper(
		&helper.StandardSleeper{},
		func(err error) bool {
			var netErr net.Error

			return errors.As(err, &netErr)
		},
		[]time.Duration{
			time.Second,
			time.Second * 3,
			time.Second * 5,
		},
		logger,
	)

	return &MetricsAgent{
		httpClient:      client,
		serverAddr:      addr,
		pollCount:       0,
		randomValue:     0,
		metrics:         make([]MemMetrics, 0),
		logger:          logger,
		retryableHelper: retryableHelper,
	}
}

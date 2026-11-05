package agent

import (
	"context"
	"errors"
	"fmt"
	"go-metrics/internal/config"
	"go-metrics/internal/model"
	"go-metrics/internal/resty/middleware"
	"net"
	"time"

	"github.com/rs/zerolog"
	"resty.dev/v3"
)

type MetricsAgent struct {
	httpClient *resty.Client
	serverAddr *config.ServerAddr
	collector  *collector
	logger     *zerolog.Logger
}

func (a *MetricsAgent) CollectMemMetrics(ctx context.Context) error {
	err := a.collector.CollectMemMetrics(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (a *MetricsAgent) CollectGopsutilMetrics(ctx context.Context) error {
	err := a.collector.CollectGopsutilMetrics(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (a *MetricsAgent) Snapshot(ctx context.Context) []*model.Metric {
	return a.collector.GetSnapshot(ctx)
}

func (a *MetricsAgent) SendReports(ctx context.Context, metrics []*model.Metric) error {
	_, err := a.httpClient.R().
		SetContext(ctx).
		SetBody(metrics).
		Post(fmt.Sprintf("http://%s/updates/", a.serverAddr.Addr))
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}

	return nil
}

func NewMetricsAgent(client *resty.Client, addr *config.ServerAddr, hashKey string, logger *zerolog.Logger) *MetricsAgent {
	client.AddContentTypeEncoder("application/json", resty.InMemoryJSONMarshal)
	client.AddRequestMiddleware(middleware.RequestGzipCompress(logger))
	client.AddRequestMiddleware(middleware.RequestHeaderHasher(hashKey))

	client.SetRetryAllowNonIdempotent(true)
	client.SetRetryCount(3)
	client.AddRetryConditions(func(_ *resty.Response, err error) bool {
		var netErr net.Error

		return errors.As(err, &netErr)
	})

	tries := []time.Duration{
		time.Second,
		time.Second * 3,
		time.Second * 5,
	}
	client.SetRetryDelayStrategy(func(r *resty.Response, err error) (time.Duration, error) {
		return tries[r.Request.Attempt-1], err
	})

	return &MetricsAgent{
		httpClient: client,
		serverAddr: addr,
		collector:  newCollector(),
		logger:     logger,
	}
}

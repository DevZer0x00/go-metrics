package agent

import (
	"fmt"
	"go-metrics/internal/config"
	"go-metrics/internal/model"
	"math/rand"

	"github.com/rs/zerolog/log"
	"resty.dev/v3"
)

type MetricsAgent struct {
	httpClient  *resty.Client
	serverAddr  *config.ServerAddr
	pollCount   int64
	randomValue float64
	metrics     []MemMetrics
}

func (a *MetricsAgent) resetAfterSend() {
	a.pollCount = 0
	a.randomValue = 0
	a.metrics = make([]MemMetrics, 0)
}

func (a *MetricsAgent) Collect() {
	a.pollCount++
	a.randomValue = rand.Float64() * 100
	a.metrics = CollectMemMetrics()
}

func (a *MetricsAgent) sendReport(metric *model.Metric) error {
	_, err := a.httpClient.R().
		SetBody(metric).
		Post(fmt.Sprintf("http://%s/update", a.serverAddr.Addr))
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}

	return nil
}

func (a *MetricsAgent) Send() {
	if a.pollCount == 0 {
		return
	}

	for _, memMetric := range a.metrics {
		metric := &model.Metric{
			ID:    memMetric.Label,
			MType: model.Gauge,
			Value: &memMetric.Value,
		}

		err := a.sendReport(metric)
		if err != nil {
			log.Err(err).Msgf("error sending metric for metric %s", memMetric.Label)
		}
	}

	err := a.sendReport(&model.Metric{
		ID:    "PollCount",
		MType: model.Counter,
		Delta: &a.pollCount,
	})
	if err != nil {
		log.
			Err(err).
			Str("metricName", "PollCount").
			Msg("error sending metric")
	}

	err = a.sendReport(&model.Metric{
		ID:    "RandomValue",
		MType: model.Gauge,
		Value: &a.randomValue,
	})
	if err != nil {
		log.
			Err(err).
			Str("metricName", "RandomValue").
			Msg("error sending metric")
	}

	a.resetAfterSend()
}

func NewMetricsAgent(client *resty.Client, addr *config.ServerAddr) *MetricsAgent {
	return &MetricsAgent{
		httpClient:  client,
		serverAddr:  addr,
		pollCount:   0,
		randomValue: 0,
		metrics:     make([]MemMetrics, 0),
	}
}

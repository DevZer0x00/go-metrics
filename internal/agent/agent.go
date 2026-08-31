package agent

import (
	"fmt"
	"go-metrics/internal/config"
	"go-metrics/internal/model"
	"io"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"
)

type MetricsAgent struct {
	httpClient  *http.Client
	serverAddr  *config.ServerAddr
	pollCount   uint64
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

func (a *MetricsAgent) sendReport(mType, label, value string) error {
	url := fmt.Sprintf("http://%s/update/%s/%s/%s", a.serverAddr.Addr, mType, label, value)

	request, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("error creating request for url %s: %w", url, err)
	}

	resp, err := a.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("error making request for url %s: %w", url, err)
	}

	_, _ = io.Copy(io.Discard, resp.Body)
	err = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error closing response body for url %s: %w", url, err)
	}

	return nil
}

func (a *MetricsAgent) Send() {
	if a.pollCount == 0 {
		return
	}

	for _, metric := range a.metrics {
		err := a.sendReport(model.Gauge, metric.Label, strconv.FormatFloat(metric.Value, 'f', -1, 64))
		if err != nil {
			log.Err(err).Msgf("error sending metric for metric %s", metric.Label)
		}
	}

	err := a.sendReport(model.Counter, "PollCount", strconv.FormatUint(a.pollCount, 10))
	if err != nil {
		log.
			Err(err).
			Str("metricName", "PollCount").
			Msg("error sending metric")
	}

	err = a.sendReport(model.Gauge, "RandomValue", strconv.FormatFloat(a.randomValue, 'f', -1, 64))
	if err != nil {
		log.
			Err(err).
			Str("metricName", "RandomValue").
			Msg("error sending metric")
	}

	a.resetAfterSend()
}

func NewMetricsAgent(client *http.Client, addr *config.ServerAddr) *MetricsAgent {
	return &MetricsAgent{
		httpClient:  client,
		serverAddr:  addr,
		pollCount:   0,
		randomValue: 0,
		metrics:     make([]MemMetrics, 0),
	}
}

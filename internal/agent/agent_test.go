package agent

import (
	"go-metrics/internal/config"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type RoundTripFunc func(req *http.Request) (*http.Response, error)

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func TestMetricAgentCollect(t *testing.T) {
	httpTestClient := newTestClient(func(req *http.Request) (*http.Response, error) {
		t.Fatal("should not be called")

		return &http.Response{}, nil
	})

	metricsAgent := NewMetricsAgent(httpTestClient, &config.ServerAddr{Addr: "localhost"})
	metricsAgent.Collect()
	assert.Equal(t, uint64(1), metricsAgent.pollCount)
	assert.Len(t, metricsAgent.metrics, 27)
	assert.NotEmpty(t, metricsAgent.randomValue)
}

func TestMetricAgentResetAfterSend(t *testing.T) {
	httpTestClient := newTestClient(func(req *http.Request) (*http.Response, error) {
		t.Fatal("should not be called")

		return &http.Response{}, nil
	})

	metricsAgent := NewMetricsAgent(httpTestClient, &config.ServerAddr{Addr: "localhost"})
	metricsAgent.Collect()
	metricsAgent.resetAfterSend()

	assert.Equal(t, uint64(0), metricsAgent.pollCount)
	assert.Len(t, metricsAgent.metrics, 0)
	assert.Empty(t, metricsAgent.randomValue)
}

func TestMetricAgentSendEmptyMetrics(t *testing.T) {
	httpTestClient := newTestClient(func(req *http.Request) (*http.Response, error) {
		t.Fatal("should not be called")

		return &http.Response{}, nil
	})

	metricsAgent := NewMetricsAgent(httpTestClient, &config.ServerAddr{Addr: "localhost"})
	metricsAgent.Send()
}

func TestMetricAgentSendMetrics(t *testing.T) {
	counter := 0

	httpTestClient := newTestClient(func(req *http.Request) (*http.Response, error) {
		counter++

		assert.Equal(t, http.MethodPost, req.Method)

		return &http.Response{}, nil
	})

	metricsAgent := NewMetricsAgent(httpTestClient, &config.ServerAddr{Addr: "localhost"})
	metricsAgent.Collect()
	metricsAgent.Send()

	assert.Equal(t, uint64(0), metricsAgent.pollCount)
	assert.Equal(t, 29, counter)
}

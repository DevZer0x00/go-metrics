package agent

import (
	"compress/gzip"
	"encoding/json"
	"go-metrics/internal/config"
	"go-metrics/internal/model"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"resty.dev/v3"
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
	client := resty.NewWithClient(httpTestClient)

	metricsAgent := NewMetricsAgent(client, &config.ServerAddr{Addr: "localhost"})
	metricsAgent.Collect()
	assert.Equal(t, int64(1), metricsAgent.pollCount)
	assert.Len(t, metricsAgent.metrics, 27)
	assert.NotEmpty(t, metricsAgent.randomValue)
}

func TestMetricAgentResetAfterSend(t *testing.T) {
	httpTestClient := newTestClient(func(req *http.Request) (*http.Response, error) {
		t.Fatal("should not be called")

		return &http.Response{}, nil
	})
	client := resty.NewWithClient(httpTestClient)

	metricsAgent := NewMetricsAgent(client, &config.ServerAddr{Addr: "localhost"})
	metricsAgent.Collect()
	metricsAgent.resetAfterSend()

	assert.Equal(t, int64(0), metricsAgent.pollCount)
	assert.Len(t, metricsAgent.metrics, 0)
	assert.Empty(t, metricsAgent.randomValue)
}

func TestMetricAgentSendEmptyMetrics(t *testing.T) {
	httpTestClient := newTestClient(func(req *http.Request) (*http.Response, error) {
		t.Fatal("should not be called")

		return &http.Response{}, nil
	})
	client := resty.NewWithClient(httpTestClient)

	metricsAgent := NewMetricsAgent(client, &config.ServerAddr{Addr: "localhost"})
	metricsAgent.Send()
}

func TestMetricAgentSendMetrics(t *testing.T) {
	counter := 0

	httpTestClient := newTestClient(func(req *http.Request) (*http.Response, error) {
		counter++

		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, "application/x-gzip", req.Header.Get("content-type"))

		gzipReader, err := gzip.NewReader(req.Body)
		require.NoError(t, err)

		defer gzipReader.Close()

		bodyBytes, err := io.ReadAll(gzipReader)
		require.NoError(t, err)

		metric := &model.Metric{}
		err = json.Unmarshal(bodyBytes, &metric)
		require.NoError(t, err)

		return &http.Response{}, nil
	})
	client := resty.NewWithClient(httpTestClient)

	metricsAgent := NewMetricsAgent(client, &config.ServerAddr{Addr: "localhost"})
	metricsAgent.Collect()
	metricsAgent.Send()

	assert.Equal(t, int64(0), metricsAgent.pollCount)
	assert.Equal(t, 29, counter)
}

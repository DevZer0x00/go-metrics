package agent

import (
	"compress/gzip"
	"encoding/json"
	"go-metrics/internal/config"
	"go-metrics/internal/model"
	"io"
	"net/http"
	"testing"

	"github.com/rs/zerolog"
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

	logger := zerolog.New(io.Discard)

	metricsAgent := NewMetricsAgent(client, &config.ServerAddr{Addr: "localhost"}, "", &logger)
	err := metricsAgent.CollectMemMetrics(t.Context())
	require.NoError(t, err)

	metrics := metricsAgent.Snapshot(t.Context())
	assert.Len(t, metrics, 29)
}

func TestMetricAgentSendBatchMetrics(t *testing.T) {
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

		metrics := make([]*model.Metric, 0)
		err = json.Unmarshal(bodyBytes, &metrics)
		require.NoError(t, err)

		return &http.Response{
			StatusCode: http.StatusOK,
		}, nil
	})
	client := resty.NewWithClient(httpTestClient)

	logger := zerolog.New(io.Discard)

	metricsAgent := NewMetricsAgent(client, &config.ServerAddr{Addr: "localhost"}, "", &logger)
	err := metricsAgent.CollectMemMetrics(t.Context())
	require.NoError(t, err)

	err = metricsAgent.SendReports(t.Context(), metricsAgent.Snapshot(t.Context()))
	require.NoError(t, err)

	assert.Equal(t, 1, counter)
}

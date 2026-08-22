package routes

import (
	"go-metrics/internal/model"
	"go-metrics/internal/repository"
	"go-metrics/internal/service"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateMetricHandlerParameters(t *testing.T) {
	tests := []struct {
		TestName   string
		Method     string
		Path       string
		StatusCode int
		Body       string
	}{
		{
			TestName:   "Metric type not allowed",
			Method:     http.MethodPost,
			Path:       "/update/badMetricType/requestTotal/23",
			StatusCode: http.StatusBadRequest,
			Body:       "Bad Request\n",
		},
		{
			TestName:   "Empty metric name 1",
			Method:     http.MethodPost,
			Path:       "/update/counter",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
		{
			TestName:   "Empty metric name 2",
			Method:     http.MethodPost,
			Path:       "/update/counter/%20/23",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
		{
			TestName:   "Invalid metric value 1",
			Method:     http.MethodPost,
			Path:       "/update/counter/metric/sdf",
			StatusCode: http.StatusBadRequest,
			Body:       "Bad Request\n",
		},
		{
			TestName:   "Invalid metric value 2",
			Method:     http.MethodPost,
			Path:       "/update/counter/metric/12.4",
			StatusCode: http.StatusBadRequest,
			Body:       "Bad Request\n",
		},
		{
			TestName:   "Invalid metric value 3",
			Method:     http.MethodPost,
			Path:       "/update/gauge/metric/sdfasdf",
			StatusCode: http.StatusBadRequest,
			Body:       "Bad Request\n",
		},
		{
			TestName:   "Correct counter",
			Method:     http.MethodPost,
			Path:       "/update/counter/metric/1",
			StatusCode: http.StatusOK,
			Body:       "",
		},
		{
			TestName:   "Correct gauge",
			Method:     http.MethodPost,
			Path:       "/update/gauge/metric/12.5",
			StatusCode: http.StatusOK,
			Body:       "",
		},
	}

	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			request := httptest.NewRequest(test.Method, test.Path, nil)
			recorder := httptest.NewRecorder()
			metricsService := service.NewMetricsService(repository.NewMemStorage())

			r := NewRouter(metricsService)
			r.ServeHTTP(recorder, request)

			response := recorder.Result()

			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %s", err)
			}

			assert.Equal(t, test.StatusCode, recorder.Code)
			assert.Equal(t, test.Body, string(body))
		})
	}
}

func TestGetMetricHandler(t *testing.T) {
	delta := int64(35)
	value := 12.33444
	metricHash := model.GetMetricHash("requestTotal")

	tests := []struct {
		TestName   string
		Method     string
		Path       string
		Metric     *model.Metric
		StatusCode int
		Value      string
	}{
		{
			TestName:   "Metric type not allowed",
			Method:     http.MethodGet,
			Path:       "/value/badMetricType/requestTotal",
			StatusCode: http.StatusBadRequest,
			Value:      "",
		},
		{
			TestName:   "Metric not found 1",
			Method:     http.MethodGet,
			Path:       "/value/counter/requestTotal",
			StatusCode: http.StatusNotFound,
			Value:      "",
		},
		{
			TestName: "Metric not found 2",
			Method:   http.MethodGet,
			Path:     "/value/gauge/requestTotal",
			Metric: &model.Metric{
				ID:    "requestTotal",
				MType: model.Counter,
				Hash:  metricHash,
			},
			StatusCode: http.StatusNotFound,
			Value:      "",
		},
		{
			TestName: "Counter metric value",
			Method:   http.MethodGet,
			Path:     "/value/counter/requestTotal",
			Metric: &model.Metric{
				ID:    "requestTotal",
				MType: model.Counter,
				Delta: &delta,
				Hash:  metricHash,
			},
			StatusCode: http.StatusOK,
			Value:      "35",
		},
		{
			TestName: "Gauge metric value",
			Method:   http.MethodGet,
			Path:     "/value/gauge/requestTotal",
			Metric: &model.Metric{
				ID:    "requestTotal",
				MType: model.Gauge,
				Value: &value,
				Hash:  metricHash,
			},
			StatusCode: http.StatusOK,
			Value:      "12.334",
		},
	}

	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			repo := repository.NewMemStorage()
			metricsService := service.NewMetricsService(repo)

			if test.Metric != nil {
				err := repo.Save(test.Metric)
				require.NoError(t, err)
			}

			request := httptest.NewRequest(test.Method, test.Path, nil)
			recorder := httptest.NewRecorder()

			r := NewRouter(metricsService)
			r.ServeHTTP(recorder, request)

			response := recorder.Result()
			defer response.Body.Close()

			body, _ := io.ReadAll(response.Body)

			assert.Equal(t, test.StatusCode, recorder.Code)

			if test.StatusCode == http.StatusOK {
				assert.Equal(t, test.Value, string(body))
			}
		})
	}
}

func TestAllMetricsHandler(t *testing.T) {
	repo := repository.NewMemStorage()
	metricsService := service.NewMetricsService(repo)

	metric1, _ := repo.GetOrRegister(model.Counter, "counter1")
	metric2, _ := repo.GetOrRegister(model.Counter, "counter2")
	metric3, _ := repo.GetOrRegister(model.Gauge, "gauge1")

	metric1.UpdateDelta(100)
	metric2.UpdateDelta(200)
	metric3.UpdateValue(12.455365436546)
	err := repo.Save(metric1)
	require.NoError(t, err)
	err = repo.Save(metric2)
	require.NoError(t, err)
	err = repo.Save(metric3)
	require.NoError(t, err)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()

	r := NewRouter(metricsService)
	r.ServeHTTP(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	_, _ = io.ReadAll(response.Body)

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "text/html; charset=utf-8", response.Header.Get("Content-Type"))
}

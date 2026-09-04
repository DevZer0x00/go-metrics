package routes

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"go-metrics/internal/model"
	"go-metrics/internal/repository"
	"go-metrics/internal/service"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateFromPathHandlerParameters(t *testing.T) {
	tests := []struct {
		TestName     string
		Method       string
		Path         string
		StatusCode   int
		ResponseBody string
	}{
		{
			TestName:     "Metric type not allowed",
			Method:       http.MethodPost,
			Path:         "/update/badMetricType/requestTotal/23",
			StatusCode:   http.StatusBadRequest,
			ResponseBody: "Bad Request\n",
		},
		{
			TestName:     "Empty metric name 1",
			Method:       http.MethodPost,
			Path:         "/update/counter",
			StatusCode:   http.StatusNotFound,
			ResponseBody: "404 page not found\n",
		},
		{
			TestName:     "Empty metric name 2",
			Method:       http.MethodPost,
			Path:         "/update/counter/%20/23",
			StatusCode:   http.StatusNotFound,
			ResponseBody: "404 page not found\n",
		},
		{
			TestName:     "Invalid metric value 1",
			Method:       http.MethodPost,
			Path:         "/update/counter/metric/sdf",
			StatusCode:   http.StatusBadRequest,
			ResponseBody: "Bad Request\n",
		},
		{
			TestName:     "Invalid metric value 2",
			Method:       http.MethodPost,
			Path:         "/update/counter/metric/12.4",
			StatusCode:   http.StatusBadRequest,
			ResponseBody: "Bad Request\n",
		},
		{
			TestName:     "Invalid metric value 3",
			Method:       http.MethodPost,
			Path:         "/update/gauge/metric/sdfasdf",
			StatusCode:   http.StatusBadRequest,
			ResponseBody: "Bad Request\n",
		},
		{
			TestName:     "Correct counter",
			Method:       http.MethodPost,
			Path:         "/update/counter/metric/1",
			StatusCode:   http.StatusOK,
			ResponseBody: "",
		},
		{
			TestName:     "Correct gauge",
			Method:       http.MethodPost,
			Path:         "/update/gauge/metric/12.5",
			StatusCode:   http.StatusOK,
			ResponseBody: "",
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

			assert.Equal(t, test.StatusCode, response.StatusCode)
			assert.Equal(t, test.ResponseBody, string(body))
		})
	}
}

func TestUpdateFromJSONHandlerFunc(t *testing.T) {
	tests := []struct {
		TestName     string
		Method       string
		ContentType  string
		RequestBody  string
		ResponseBody string
		StatusCode   int
	}{
		{
			TestName:     "Invalid Content Type",
			Method:       http.MethodPost,
			RequestBody:  "",
			StatusCode:   http.StatusBadRequest,
			ResponseBody: "Bad Request\n",
			ContentType:  "text/plain; charset=utf-8",
		},
		{
			TestName:     "Invalid Json",
			Method:       http.MethodPost,
			RequestBody:  `{"id": "test", "type": "gauge", "value: 12.14124}`,
			StatusCode:   http.StatusBadRequest,
			ResponseBody: "Bad Request\n",
		},
		{
			TestName:     "Metric type not allowed",
			Method:       http.MethodPost,
			RequestBody:  `{"id": "test", "type": "badType", "value": 1214124}`,
			StatusCode:   http.StatusBadRequest,
			ResponseBody: "Bad Request\n",
		},
		{
			TestName:     "Empty metric name 1",
			Method:       http.MethodPost,
			RequestBody:  `{"id": "    ", "type": "counter", "value": 1214124}`,
			StatusCode:   http.StatusNotFound,
			ResponseBody: "404 page not found\n",
		},
		{
			TestName:     "Empty metric name 2",
			Method:       http.MethodPost,
			RequestBody:  `{"type": "badType", "value": 1214124}`,
			StatusCode:   http.StatusNotFound,
			ResponseBody: "404 page not found\n",
		},
		{
			TestName:     "Invalid metric value 1",
			Method:       http.MethodPost,
			RequestBody:  `{"id": "test", "type": "counter", "delta": "sdsafg"}`,
			StatusCode:   http.StatusBadRequest,
			ResponseBody: "Bad Request\n",
		},
		{
			TestName:     "Invalid metric value 2",
			Method:       http.MethodPost,
			RequestBody:  `{"id": "test", "type": "counter", "delta": 12.14124}`,
			StatusCode:   http.StatusBadRequest,
			ResponseBody: "Bad Request\n",
		},
		{
			TestName:     "Invalid metric value 3",
			Method:       http.MethodPost,
			RequestBody:  `{"id": "test", "type": "gauge", "value": "sdgsdg"}`,
			StatusCode:   http.StatusBadRequest,
			ResponseBody: "Bad Request\n",
		},
		{
			TestName:     "Invalid metric value for metric type counter",
			Method:       http.MethodPost,
			RequestBody:  `{"id": "test", "type": "counter", "value": 12.44}`,
			StatusCode:   http.StatusBadRequest,
			ResponseBody: "Bad Request\n",
		},
		{
			TestName:     "Invalid metric value for metric type gauge",
			Method:       http.MethodPost,
			RequestBody:  `{"id": "test", "type": "gauge", "delta": 12}`,
			StatusCode:   http.StatusBadRequest,
			ResponseBody: "Bad Request\n",
		},
		{
			TestName:     "Correct counter",
			Method:       http.MethodPost,
			RequestBody:  `{"id": "test", "type": "counter", "delta": 1214124}`,
			StatusCode:   http.StatusOK,
			ResponseBody: "",
		},
		{
			TestName:     "Correct gauge",
			Method:       http.MethodPost,
			RequestBody:  `{"id": "test", "type": "gauge", "value": 12.14124}`,
			StatusCode:   http.StatusOK,
			ResponseBody: "",
		},
	}

	for _, compressed := range []bool{true, false} {
		for _, test := range tests {
			testName := fmt.Sprintf("%s compressed: %t", test.TestName, compressed)

			t.Run(testName, func(t *testing.T) {
				if compressed {
					var buffer bytes.Buffer
					gzipWriter, _ := gzip.NewWriterLevel(&buffer, gzip.BestCompression)
					_, _ = gzipWriter.Write([]byte(test.RequestBody))
					_ = gzipWriter.Close()
					test.RequestBody = buffer.String()
				}

				request := httptest.NewRequest(test.Method, "/update", strings.NewReader(test.RequestBody))
				if len(test.ContentType) > 0 {
					request.Header.Set("Content-Type", test.ContentType)
				} else {
					if compressed {
						request.Header.Set("Content-Encoding", "gzip")
						request.Header.Set("Content-Type", "application/x-gzip")
					} else {
						request.Header.Set("Content-Type", "application/json")
					}
				}

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
				assert.Equal(t, test.ResponseBody, string(body))
			})
		}
	}
}

func TestGetMetricHandler(t *testing.T) {
	delta := int64(35)
	value := 12.33000
	metricHash := model.GetMetricHash("requestTotal")

	tests := []struct {
		TestName   string
		Path       string
		Metric     *model.Metric
		StatusCode int
		Value      string
	}{
		{
			TestName:   "Metric type not allowed",
			Path:       "/value/badMetricType/requestTotal",
			StatusCode: http.StatusBadRequest,
			Value:      "",
		},
		{
			TestName:   "Metric not found 1",
			Path:       "/value/counter/requestTotal",
			StatusCode: http.StatusNotFound,
			Value:      "",
		},
		{
			TestName: "Metric not found 2",
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
			Path:     "/value/gauge/requestTotal",
			Metric: &model.Metric{
				ID:    "requestTotal",
				MType: model.Gauge,
				Value: &value,
				Hash:  metricHash,
			},
			StatusCode: http.StatusOK,
			Value:      "12.33",
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

			request := httptest.NewRequest(http.MethodGet, test.Path, nil)
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

func TestGetMetricValueHandler(t *testing.T) {
	delta := int64(35)
	value := 12.33444
	metricHash := model.GetMetricHash("requestTotal")

	tests := []struct {
		TestName       string
		Metric         *model.Metric
		RequestBody    string
		StatusCode     int
		ResponseBody   string
		AcceptEncoding string
	}{
		{
			TestName:    "Empty json",
			RequestBody: ``,
			StatusCode:  http.StatusBadRequest,
		},
		{
			TestName:    "Metric not found 1",
			RequestBody: `{"id": "test", "type": "requestTotal"}`,
			StatusCode:  http.StatusNotFound,
		},
		{
			TestName: "Metric not found 2",
			Metric: &model.Metric{
				ID:    "requestTotal",
				MType: model.Counter,
				Hash:  metricHash,
			},
			RequestBody: `{"id": "test", "type": "gauge"}`,
			StatusCode:  http.StatusNotFound,
		},
		{
			TestName: "Counter metric value",
			Metric: &model.Metric{
				ID:    "requestTotal",
				MType: model.Counter,
				Delta: &delta,
				Hash:  metricHash,
			},
			RequestBody:  `{"id": "requestTotal", "type": "counter"}`,
			StatusCode:   http.StatusOK,
			ResponseBody: `{"id":"requestTotal","type":"counter","delta":35}`,
		},
		{
			TestName: "Counter metric value gzipped",
			Metric: &model.Metric{
				ID:    "requestTotal",
				MType: model.Counter,
				Delta: &delta,
				Hash:  metricHash,
			},
			RequestBody:    `{"id": "requestTotal", "type": "counter"}`,
			StatusCode:     http.StatusOK,
			ResponseBody:   `{"id":"requestTotal","type":"counter","delta":35}`,
			AcceptEncoding: "gzip",
		},
		{
			TestName: "Gauge metric value",
			Metric: &model.Metric{
				ID:    "requestTotal",
				MType: model.Gauge,
				Value: &value,
				Hash:  metricHash,
			},
			RequestBody:  `{"id": "requestTotal", "type": "gauge"}`,
			StatusCode:   http.StatusOK,
			ResponseBody: `{"id":"requestTotal","type":"gauge","value":12.33444}`,
		},
		{
			TestName: "Gauge metric value",
			Metric: &model.Metric{
				ID:    "requestTotal",
				MType: model.Gauge,
				Value: &value,
				Hash:  metricHash,
			},
			RequestBody:    `{"id": "requestTotal", "type": "gauge"}`,
			StatusCode:     http.StatusOK,
			ResponseBody:   `{"id":"requestTotal","type":"gauge","value":12.33444}`,
			AcceptEncoding: "gzip",
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

			request := httptest.NewRequest(http.MethodPost, "/value", strings.NewReader(test.RequestBody))
			request.Header.Set("Content-Type", "application/json")
			if test.AcceptEncoding != "" {
				request.Header.Set("Accept-Encoding", test.AcceptEncoding)
			}

			recorder := httptest.NewRecorder()

			r := NewRouter(metricsService)
			r.ServeHTTP(recorder, request)

			response := recorder.Result()
			defer response.Body.Close()

			body, _ := io.ReadAll(response.Body)
			if response.Header.Get("Content-Encoding") == "gzip" {
				gzipReader, err := gzip.NewReader(bytes.NewReader(body))
				require.NoError(t, err)

				body, err = io.ReadAll(gzipReader)
				require.NoError(t, err)
			}

			assert.Equal(t, test.StatusCode, recorder.Code)

			if test.StatusCode == http.StatusOK {
				assert.Equal(t, test.ResponseBody, strings.TrimSpace(string(body)))
				assert.Equal(t, "application/json", response.Header.Get("Content-Type"))
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

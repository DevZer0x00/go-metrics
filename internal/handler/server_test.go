package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricUpdateHandlerParameters(t *testing.T) {
	tests := []struct {
		TestName   string
		Method     string
		Path       string
		StatusCode int
		Body       string
	}{
		{
			TestName:   "Not allowed method",
			Method:     http.MethodGet,
			Path:       "/update/counter/requestTotal/23",
			StatusCode: http.StatusMethodNotAllowed,
			Body:       "Method Not Allowed\n",
		},
		{
			TestName:   "Metric type not allowed",
			Method:     http.MethodPost,
			Path:       "/update/badMetricType/requestTotal/23",
			StatusCode: http.StatusBadRequest,
			Body:       "Bad request\n",
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
			TestName:   "Empty metric value",
			Method:     http.MethodPost,
			Path:       "/update/counter/metric",
			StatusCode: http.StatusBadRequest,
			Body:       "Bad request\n",
		},
		{
			TestName:   "Invalid metric value 1",
			Method:     http.MethodPost,
			Path:       "/update/counter/metric/sdf",
			StatusCode: http.StatusBadRequest,
			Body:       "Bad request\n",
		},
		{
			TestName:   "Invalid metric value 2",
			Method:     http.MethodPost,
			Path:       "/update/counter/metric/12.4",
			StatusCode: http.StatusBadRequest,
			Body:       "Bad request\n",
		},
		{
			TestName:   "Invalid metric value 3",
			Method:     http.MethodPost,
			Path:       "/update/gauge/metric/sdfasdf",
			StatusCode: http.StatusBadRequest,
			Body:       "Bad request\n",
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

			mux := NewRouter()
			mux.ServeHTTP(recorder, request)

			response := recorder.Result()

			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %s", err)
			}

			assert.Equal(t, test.StatusCode, recorder.Code)
			assert.Equal(t, string(body), test.Body)
		})
	}
}

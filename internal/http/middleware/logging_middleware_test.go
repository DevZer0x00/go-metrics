package middleware

import (
	"bytes"
	"encoding/json"
	"go-metrics/internal/config"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggingMiddleware(t *testing.T) {
	var buffer bytes.Buffer

	router := chi.NewRouter()
	router.Use(LoggingMiddleware())
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)

		_, _ = w.Write([]byte("Hello World"))
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(router)
	defer server.Close()

	config.InitLog(&buffer)

	resp, err := http.Get(server.URL)
	require.NoError(t, err)

	err = resp.Body.Close()
	require.NoError(t, err)

	var logData struct {
		RequestMethod string  `json:"requestMethod,required"`
		RequestUri    string  `json:"requestUri,required"`
		ResponseTime  float64 `json:"responseTime,required"`
		ResponseCode  int     `json:"responseCode,required"`
		ResponseSize  uint64  `json:"responseSize,required"`
	}

	err = json.Unmarshal(buffer.Bytes(), &logData)
	require.NoError(t, err)

	assert.Equal(t, "GET", logData.RequestMethod)
	assert.Equal(t, "/", logData.RequestUri)
	assert.NotEmpty(t, logData.ResponseTime)
	assert.Equal(t, http.StatusOK, logData.ResponseCode)
}

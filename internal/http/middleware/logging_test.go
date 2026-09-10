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
	logger := config.InitLog(&buffer)

	router := chi.NewRouter()
	router.Use(LoggingMiddleware(logger))
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)

		_, _ = w.Write([]byte("Hello World"))
	})

	server := httptest.NewServer(router)
	defer server.Close()

	resp, err := http.Get(server.URL)
	require.NoError(t, err)

	err = resp.Body.Close()
	require.NoError(t, err)

	var logData struct {
		RequestMethod string  `json:"requestMethod"`
		RequestURI    string  `json:"requestURI"`
		ResponseTime  float64 `json:"responseTime"`
	}

	err = json.Unmarshal(buffer.Bytes(), &logData)
	require.NoError(t, err)

	assert.Equal(t, "GET", logData.RequestMethod)
	assert.Equal(t, "/", logData.RequestURI)
	assert.NotEmpty(t, logData.ResponseTime)
}

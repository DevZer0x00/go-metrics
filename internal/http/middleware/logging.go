package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.status = statusCode
}

func (w *loggingResponseWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}

	size, err := w.ResponseWriter.Write(data)

	return size, err
}

func LoggingMiddleware(logger *zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()

			writer := &loggingResponseWriter{
				ResponseWriter: w,
				status:         0,
			}

			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "cant read body", http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			next.ServeHTTP(writer, r)
			endTime := time.Since(startTime)

			logger.Info().
				Str("requestMethod", r.Method).
				Str("requestUri", r.URL.RequestURI()).
				Dur("responseTime", endTime).
				Msg("Incoming request")
		})
	}
}

package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *struct {
		statusCode int
		length     uint64
	}
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.responseData.statusCode = statusCode
}

func (w *loggingResponseWriter) Write(data []byte) (int, error) {
	size, err := w.ResponseWriter.Write(data)
	w.responseData.length += uint64(len(data))

	return size, err
}

func LoggingMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()

			writer := &loggingResponseWriter{
				ResponseWriter: w,
				responseData: new(struct {
					statusCode int
					length     uint64
				}),
			}
			next.ServeHTTP(writer, r)
			endTime := time.Since(startTime)

			log.Info().
				Str("requestMethod", r.Method).
				Str("requestUri", r.URL.RequestURI()).
				Dur("responseTime", endTime).
				Int("responseCode", writer.responseData.statusCode).
				Uint64("responseSize", writer.responseData.length).
				Msg("Incoming request")
		})
	}
}

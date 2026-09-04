package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

func GzipEncodedMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Content-Encoding") == "gzip" {
				gzipReader, err := gzip.NewReader(r.Body)
				if err != nil {
					log.
						Error().
						Err(err).
						Msg("error creating gzip reader")
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}

				defer gzipReader.Close()

				bodyBuffer := new(bytes.Buffer)

				_, err = io.Copy(bodyBuffer, gzipReader)
				if err != nil {
					log.
						Error().
						Err(err).
						Msg("error reading body")
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}

				bodyType := http.DetectContentType(bodyBuffer.Bytes())
				if bodyType != "text/plain; charset=utf-8" {
					http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
					return
				}

				r.Body = io.NopCloser(bodyBuffer)
				r.Header.Set("Content-Type", "application/json")
			}

			next.ServeHTTP(w, r)
		})
	}
}

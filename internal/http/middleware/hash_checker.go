package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

type hasherResponseWriter struct {
	http.ResponseWriter
	hashKey string
}

func (w *hasherResponseWriter) Write(data []byte) (int, error) {
	hasher := hmac.New(sha256.New, []byte(w.hashKey))
	hasher.Write(data)
	w.Header().Set("HashSHA256", hex.EncodeToString(hasher.Sum(nil)))

	return w.ResponseWriter.Write(data)
}

func HashCheckerMiddleware(hashKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestHash := r.Header.Get("HashSHA256")

			if hashKey == "" || requestHash == "" {
				next.ServeHTTP(w, r)
				return
			}

			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "cant read body", http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			hasher := hmac.New(sha256.New, []byte(hashKey))
			hasher.Write(bodyBytes)

			if requestHash != hex.EncodeToString(hasher.Sum(nil)) {
				http.Error(w, "hash check fail", http.StatusBadRequest)
			}

			next.ServeHTTP(
				&hasherResponseWriter{
					ResponseWriter: w,
					hashKey:        hashKey,
				},
				r,
			)
		})
	}
}

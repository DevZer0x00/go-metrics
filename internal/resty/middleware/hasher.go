package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"resty.dev/v3"
)

func RequestHeaderHasher(hashKey string) func(_ *resty.Client, request *resty.Request) error {
	return func(_ *resty.Client, request *resty.Request) error {
		if hashKey == "" {
			return nil
		}

		var bodyBytes []byte

		switch b := request.Body.(type) {
		case []byte:
			bodyBytes = b
		case string:
			bodyBytes = []byte(b)
		default:
			return nil
		}

		hasher := hmac.New(sha256.New, []byte(hashKey))
		hasher.Write(bodyBytes)

		request.SetHeader("HashSHA256", hex.EncodeToString(hasher.Sum(nil)))

		return nil
	}
}

package middleware

import (
	"io"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"resty.dev/v3"
)

func TestRequestGzipCompress(t *testing.T) {
	tests := []struct {
		Name                    string
		Body                    any
		ExpectedContentEncoding string
	}{
		{
			Name:                    "string compress",
			Body:                    "sdgfasdgadgasdg",
			ExpectedContentEncoding: "gzip",
		},
		{
			Name:                    "byte compress",
			Body:                    []byte("sdgfasdgadgasdg"),
			ExpectedContentEncoding: "gzip",
		},
		{
			Name:                    "metric model compress",
			Body:                    "sdgfasdgadgasdg",
			ExpectedContentEncoding: "gzip",
		},
		{
			Name:                    "not compress",
			Body:                    120,
			ExpectedContentEncoding: "",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := resty.New()
			request := client.R().SetBody(test.Body)

			logger := zerolog.New(io.Discard)
			err := RequestGzipCompress(&logger)(client, request)
			assert.NoError(t, err)

			assert.Equal(t, test.ExpectedContentEncoding, request.Header.Get("Content-Encoding"))
		})
	}
}

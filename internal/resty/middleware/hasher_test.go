package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"resty.dev/v3"
)

func TestRequestHeaderHasher(t *testing.T) {
	tests := []struct {
		Name      string
		Body      any
		HashKey   string
		HasHeader bool
	}{
		{
			Name:      "without hash key",
			Body:      []byte("sdgfasdgadgasdg"),
			HashKey:   "",
			HasHeader: false,
		},
		{
			Name:      "byte body with hash key",
			Body:      []byte("sdgfasdgadgasdg"),
			HashKey:   "test",
			HasHeader: true,
		},
		{
			Name:      "string body with hash key",
			Body:      "sdgfasdgadgasdg",
			HashKey:   "test",
			HasHeader: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := resty.New()
			request := client.R().SetBody(test.Body)

			err := RequestHeaderHasher(test.HashKey)(client, request)
			assert.NoError(t, err)

			header := request.Header.Get("HashSHA256")
			assert.Equal(t, test.HasHeader, header != "")
		})
	}
}

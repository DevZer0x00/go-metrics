package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseServerCliOptions(t *testing.T) {
	tests := []struct {
		TestName  string
		Arguments []string
		Config    *ServerConfig
		HasError  bool
	}{
		{
			TestName:  "Default values",
			Arguments: []string{},
			Config: &ServerConfig{
				Addr: &ServerAddr{
					Addr: "localhost:8080",
				},
			},
			HasError: false,
		},
		{
			TestName: "Set options",
			Arguments: []string{
				"-a",
				"127.0.0.1:8090",
			},
			Config: &ServerConfig{
				Addr: &ServerAddr{
					Addr: "127.0.0.1:8090",
				},
			},
			HasError: false,
		},
		{
			TestName: "Bad options",
			Arguments: []string{
				"-ab",
				"127.0.0.1:8090",
			},
			Config:   nil,
			HasError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			config, err := ParseServerCliOptions(test.Arguments)
			require.Equal(t, test.HasError, err != nil, err)
			assert.Equal(t, test.Config, config)
		})
	}
}

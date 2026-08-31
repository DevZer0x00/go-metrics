package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseServerCliOptions(t *testing.T) {
	tests := []struct {
		TestName     string
		Environments []string
		Arguments    []string
		Config       *ServerConfig
		HasError     bool
	}{
		{
			TestName:     "Default values",
			Environments: []string{},
			Arguments:    []string{},
			Config: &ServerConfig{
				Addr: &ServerAddr{
					Addr: "localhost:8080",
				},
			},
			HasError: false,
		},
		{
			TestName:     "Set options",
			Environments: []string{},
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
			TestName:     "Bad options",
			Environments: []string{},
			Arguments: []string{
				"-ab",
				"127.0.0.1:8090",
			},
			Config:   nil,
			HasError: true,
		},
		{
			TestName: "Env options override arguments",
			Environments: []string{
				"ADDRESS=127.0.0.1",
			},
			Arguments: []string{
				"-a",
				"127.0.0.1:8090",
			},
			Config: &ServerConfig{
				Addr: &ServerAddr{
					Addr: "127.0.0.1",
				},
			},
			HasError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			config, err := ParseServerOptions(test.Environments, test.Arguments)
			require.Equal(t, test.HasError, err != nil, err)
			assert.Equal(t, test.Config, config)
		})
	}
}

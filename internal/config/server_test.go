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
				Persistence: &ServerPersistence{
					Interval: 300,
					Storage: &ServerPersistenceStorage{
						StorageFilePath: "./server.db",
						RestoreOnStart:  false,
					},
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
				"-i",
				"1000",
				"-f",
				"/tmp/server.db",
				"-r",
				"1",
			},
			Config: &ServerConfig{
				Addr: &ServerAddr{
					Addr: "127.0.0.1:8090",
				},
				Persistence: &ServerPersistence{
					Interval: 1000,
					Storage: &ServerPersistenceStorage{
						StorageFilePath: "/tmp/server.db",
						RestoreOnStart:  true,
					},
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
				"RESTORE_ON_START=true",
			},
			Arguments: []string{
				"-a",
				"127.0.0.1:8090",
				"-r",
				"false",
			},
			Config: &ServerConfig{
				Addr: &ServerAddr{
					Addr: "127.0.0.1",
				},
				Persistence: &ServerPersistence{
					Interval: 300,
					Storage: &ServerPersistenceStorage{
						StorageFilePath: "./server.db",
						RestoreOnStart:  true,
					},
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

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
					File: &ServerFilePersistence{
						StorageFilePath: "",
						RestoreOnStart:  false,
					},
					Database: &ServerDatabasePersistense{
						DSN: "",
					},
					Interval: 300,
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
				"-d",
				"localhost",
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
					File: &ServerFilePersistence{
						StorageFilePath: "/tmp/server.db",
						RestoreOnStart:  true,
					},
					Database: &ServerDatabasePersistense{
						DSN: "localhost",
					},
					Interval: 1000,
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
				"DATABASE_DSN=localhost1",
			},
			Arguments: []string{
				"-a",
				"127.0.0.1:8090",
				"-r",
				"false",
				"-d",
				"localhost",
			},
			Config: &ServerConfig{
				Addr: &ServerAddr{
					Addr: "127.0.0.1",
				},
				Persistence: &ServerPersistence{
					File: &ServerFilePersistence{
						StorageFilePath: "",
						RestoreOnStart:  true,
					},
					Database: &ServerDatabasePersistense{
						DSN: "localhost1",
					},
					Interval: 300,
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

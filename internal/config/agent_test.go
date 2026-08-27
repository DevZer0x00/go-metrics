package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAgentCliOptions(t *testing.T) {
	tests := []struct {
		TestName  string
		Arguments []string
		Config    *AgentConfig
		HasError  bool
	}{
		{
			TestName:  "Default values",
			Arguments: []string{},
			Config: &AgentConfig{
				ServerAddr: &ServerAddr{
					Addr: "localhost:8080",
				},
				Poll: &PollConfig{
					Interval: 2,
				},
				Report: &ReportConfig{
					Interval: 10,
				},
			},
			HasError: false,
		},
		{
			TestName: "Set options",
			Arguments: []string{
				"-a",
				"127.0.0.1:8090",
				"-r",
				"100",
				"-p",
				"200",
			},
			Config: &AgentConfig{
				ServerAddr: &ServerAddr{
					Addr: "127.0.0.1:8090",
				},
				Poll: &PollConfig{
					Interval: 200,
				},
				Report: &ReportConfig{
					Interval: 100,
				},
			},
			HasError: false,
		},
		{
			TestName: "Bad options",
			Arguments: []string{
				"-ab",
				"127.0.0.1:8090",
				"-r",
				"100",
				"-p",
				"200",
			},
			Config:   nil,
			HasError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			config, err := ParseAgentCliOptions(test.Arguments)
			require.Equal(t, test.HasError, err != nil, err)
			assert.Equal(t, test.Config, config)
		})
	}
}

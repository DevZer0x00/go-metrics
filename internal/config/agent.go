package config

import "flag"

type PollConfig struct {
	Interval uint64
}

type ReportConfig struct {
	Interval uint64
}

type AgentConfig struct {
	ServerAddr *ServerAddr
	Poll       *PollConfig
	Report     *ReportConfig
}

func ParseAgentCliOptions(arguments []string) (*AgentConfig, error) {
	cfg := &AgentConfig{
		ServerAddr: &ServerAddr{
			Addr: "localhost:8080",
		},
		Poll: &PollConfig{
			Interval: 2,
		},
		Report: &ReportConfig{
			Interval: 10,
		},
	}

	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.StringVar(&cfg.ServerAddr.Addr, "a", cfg.ServerAddr.Addr, "address to send metrics")
	fs.Uint64Var(&cfg.Report.Interval, "r", cfg.Report.Interval, "how often to report metrics (second)")
	fs.Uint64Var(&cfg.Poll.Interval, "p", cfg.Poll.Interval, "how often to poll metrics (second)")

	err := fs.Parse(arguments)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

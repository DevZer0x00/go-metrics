package config

import "flag"

type ServerAddr struct {
	Addr string
}

type ServerConfig struct {
	Addr *ServerAddr
}

func ParseServerCliOptions(arguments []string) (*ServerConfig, error) {
	cfg := &ServerConfig{
		Addr: &ServerAddr{
			Addr: "localhost:8080",
		},
	}

	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.StringVar(&cfg.Addr.Addr, "a", cfg.Addr.Addr, "address to listen on")

	err := fs.Parse(arguments)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

package config

import (
	"flag"

	"github.com/caarlos0/env/v11"
)

type ServerAddr struct {
	Addr string `env:"ADDRESS"`
}

type ServerConfig struct {
	Addr *ServerAddr
}

func ParseServerOptions(environments []string, arguments []string) (*ServerConfig, error) {
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

	err = env.ParseWithOptions(cfg, env.Options{
		Environment: toMap(environments),
	})

	return cfg, nil
}

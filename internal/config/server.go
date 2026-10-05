package config

import (
	"flag"
	envUtil "go-metrics/pkg/env"

	"github.com/caarlos0/env/v11"
)

type ServerAddr struct {
	Addr string `env:"ADDRESS"`
}

type ServerPersistence struct {
	File     *ServerFilePersistence
	Database *ServerDatabasePersistense
	Interval uint64 `env:"STORE_INTERVAL"`
}

type ServerFilePersistence struct {
	StorageFilePath string `env:"FILE_STORAGE_PATH"`
	RestoreOnStart  bool   `env:"RESTORE"`
}

type ServerDatabasePersistense struct {
	DSN string `env:"DATABASE_DSN"`
}

type ServerConfig struct {
	Addr        *ServerAddr
	Persistence *ServerPersistence
}

func ParseServerOptions(environments []string, arguments []string) (*ServerConfig, error) {
	cfg := &ServerConfig{
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
	}

	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.StringVar(&cfg.Addr.Addr, "a", cfg.Addr.Addr, "address to listen on")

	fs.Uint64Var(&cfg.Persistence.Interval, "i", cfg.Persistence.Interval, "flush metrics on disk interval in seconds")
	fs.StringVar(&cfg.Persistence.File.StorageFilePath, "f", cfg.Persistence.File.StorageFilePath, "file path to store metrics")
	fs.BoolVar(&cfg.Persistence.File.RestoreOnStart, "r", cfg.Persistence.File.RestoreOnStart, "restore metrics on start")

	fs.StringVar(&cfg.Persistence.Database.DSN, "d", cfg.Persistence.Database.DSN, "database connection string")

	err := fs.Parse(arguments)
	if err != nil {
		return nil, err
	}

	err = env.ParseWithOptions(cfg, env.Options{
		Environment: envUtil.ToMap(environments),
	})

	return cfg, err
}

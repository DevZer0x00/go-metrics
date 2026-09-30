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
	Interval uint64 `env:"STORE_INTERVAL"`
	Storage  *ServerPersistenceStorage
}

type ServerPersistenceStorage struct {
	StorageFilePath string `env:"FILE_STORAGE_PATH"`
	RestoreOnStart  bool   `env:"RESTORE"`
}

type Database struct {
	DSN string `env:"DATABASE_DSN"`
}

type ServerConfig struct {
	Addr        *ServerAddr
	Persistence *ServerPersistence
	Database    *Database
}

func ParseServerOptions(environments []string, arguments []string) (*ServerConfig, error) {
	cfg := &ServerConfig{
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
		Database: &Database{},
	}

	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.StringVar(&cfg.Addr.Addr, "a", cfg.Addr.Addr, "address to listen on")

	fs.Uint64Var(&cfg.Persistence.Interval, "i", cfg.Persistence.Interval, "flush metrics on disk interval in seconds")
	fs.StringVar(&cfg.Persistence.Storage.StorageFilePath, "f", cfg.Persistence.Storage.StorageFilePath, "file path to store metrics")
	fs.BoolVar(&cfg.Persistence.Storage.RestoreOnStart, "r", cfg.Persistence.Storage.RestoreOnStart, "restore metrics on start")

	fs.StringVar(&cfg.Database.DSN, "d", cfg.Database.DSN, "database connection string")

	err := fs.Parse(arguments)
	if err != nil {
		return nil, err
	}

	err = env.ParseWithOptions(cfg, env.Options{
		Environment: envUtil.ToMap(environments),
	})

	return cfg, err
}

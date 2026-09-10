package main

import (
	"go-metrics/internal/app"
	"go-metrics/internal/config"
	"os"
)

func main() {
	logger := config.InitLog(os.Stdout)

	if err := app.RunServer(os.Environ(), os.Args[1:], logger); err != nil {
		logger.
			Fatal().
			Err(err).
			Msg("failed to start server")
	}
}

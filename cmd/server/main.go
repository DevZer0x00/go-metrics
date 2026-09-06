package main

import (
	"go-metrics/internal/app"
	"os"

	"github.com/rs/zerolog/log"
)

func main() {
	if err := app.RunServer(os.Environ(), os.Args[1:]); err != nil {
		log.
			Fatal().
			Err(err).
			Msg("failed to start server")
	}
}

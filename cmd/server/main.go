package main

import (
	"context"
	"errors"
	"go-metrics/internal/app"
	"go-metrics/internal/config"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger := config.InitLog(os.Stdout)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.RunServer(ctx, os.Environ(), os.Args[1:], logger); err != nil && !errors.Is(err, context.Canceled) {
		logger.
			Fatal().
			Err(err).
			Msg("failed to start server")
	}
}

package main

import (
	"context"
	"go-metrics/internal/agent"
	"go-metrics/internal/config"
	"os"
	"os/signal"
	"syscall"
	"time"

	"resty.dev/v3"
)

func main() {
	logger := config.InitLog(os.Stdout)

	cfg, err := config.ParseAgentOptions(os.Environ(), os.Args[1:])
	if err != nil {
		logger.
			Fatal().
			Err(err).
			Msg("error parsing agent options")
	}

	logger.Info().Any("cfg", cfg).Msg("starting agent")

	client := resty.New()
	defer client.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	agentService := agent.NewMetricsAgent(client, cfg.ServerAddr, logger)

	collectTick := time.NewTicker(time.Duration(cfg.Poll.Interval) * time.Second)
	defer collectTick.Stop()

	reportTick := time.NewTicker(time.Duration(cfg.Report.Interval) * time.Second)
	defer reportTick.Stop()

	agentService.Collect()

	for {
		select {
		case <-ctx.Done():
			return
		case <-collectTick.C:
			err = agentService.Collect()
			if err != nil {
				logger.
					Fatal().
					Err(err).
					Msg("collect error")
			}
		case <-reportTick.C:
			agentService.SendBatch(ctx)
		}
	}
}

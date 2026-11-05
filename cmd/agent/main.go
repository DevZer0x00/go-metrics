package main

import (
	"context"
	"go-metrics/internal/agent"
	"go-metrics/internal/config"
	"go-metrics/internal/model"
	"os"
	"os/signal"
	"sync"
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

	logger.
		Info().
		Str("agentAddr", cfg.ServerAddr.Addr).
		Uint64("pollInterval", cfg.Poll.Interval).
		Uint64("reportInterval", cfg.Report.Interval).
		Str("hashKey", cfg.HeaderHashKey).
		Msg("starting agent")

	client := resty.New()
	defer client.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	agentService := agent.NewMetricsAgent(client, cfg.ServerAddr, cfg.HeaderHashKey, logger)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		collectTick := time.NewTicker(time.Duration(cfg.Poll.Interval) * time.Second)
		defer collectTick.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-collectTick.C:
				err := agentService.CollectMemMetrics(ctx)
				if err != nil {
					logger.
						Fatal().
						Err(err).
						Msg("collect error")
				}
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		collectTick := time.NewTicker(time.Duration(cfg.Poll.Interval) * time.Second)
		defer collectTick.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-collectTick.C:
				err := agentService.CollectGopsutilMetrics(ctx)
				if err != nil {
					logger.
						Fatal().
						Err(err).
						Msg("collect gopsutil error")
				}
			}
		}
	}()

	jobs := make(chan []*model.Metric, cfg.RateLimit)

	for i := uint64(0); i < cfg.RateLimit; i++ {
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case metrics, ok := <-jobs:
					if !ok {
						return
					}

					err := agentService.SendReports(ctx, metrics)
					if err != nil {
						logger.
							Fatal().
							Err(err).
							Msg("send reports")
					}
				}
			}
		}()
	}

	reportTick := time.NewTicker(time.Duration(cfg.Report.Interval) * time.Second)
	defer reportTick.Stop()

	for {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return
		case <-reportTick.C:
			jobs <- agentService.Snapshot(ctx)
		}
	}
}

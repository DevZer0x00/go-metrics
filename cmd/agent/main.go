package main

import (
	"go-metrics/internal/agent"
	"go-metrics/internal/config"
	"os"
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

	var timer uint64 = 0

	client := resty.New()
	defer client.Close()

	agentService := agent.NewMetricsAgent(client, cfg.ServerAddr, logger)

	for {
		if timer%cfg.Poll.Interval == 0 {
			err = agentService.Collect()
			if err != nil {
				logger.
					Fatal().
					Err(err).
					Msg("collect error")
			}
		}

		if timer != 0 && timer%cfg.Report.Interval == 0 {
			agentService.Send()
		}

		time.Sleep(time.Second)
		timer++
	}
}

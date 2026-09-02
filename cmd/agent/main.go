package main

import (
	"go-metrics/internal/agent"
	"go-metrics/internal/config"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"resty.dev/v3"
)

func main() {
	config.InitLog(os.Stdout)

	cfg, err := config.ParseAgentOptions(os.Environ(), os.Args[1:])
	if err != nil {
		log.
			Fatal().
			Err(err).
			Msg("error parsing agent options")
	}

	var timer uint64 = 0

	client := resty.New()
	defer client.Close()

	client.AddContentTypeEncoder("application/json", resty.InMemoryJSONMarshal)

	agentService := agent.NewMetricsAgent(client, cfg.ServerAddr)

	for {
		if timer%cfg.Poll.Interval == 0 {
			agentService.Collect()
		}

		if timer != 0 && timer%cfg.Report.Interval == 0 {
			agentService.Send()
		}

		time.Sleep(time.Second)
		timer++
	}
}

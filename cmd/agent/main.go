package main

import (
	"fmt"
	"go-metrics/internal/agent"
	"go-metrics/internal/config"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	cfg, err := config.ParseAgentCliOptions(os.Args[1:])
	if err != nil {
		log.Fatalln(fmt.Errorf("error parsing agent cli options: %w", err))
	}

	var timer uint64 = 0

	client := &http.Client{}
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

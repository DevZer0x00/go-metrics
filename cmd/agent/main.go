package main

import (
	"fmt"
	"go-metrics/internal/agent"
	"math/rand/v2"
	"net/http"
	"slices"
	"time"
)

func main() {
	const (
		port           = 8080
		host           = "http://127.0.0.1"
		pollInterval   = 2 * time.Second
		reportInterval = 10 * time.Second
	)

	var PollCount uint64 = 0
	var sleepBeforeSend uint64 = 0

	client := &http.Client{}

	toSend := make([]string, 0)

	for {
		memMetrics := agent.CollectMemMetrics()
		urlsLength := len(memMetrics) + 2
		urls := make([]string, urlsLength)

		for index, metrics := range memMetrics {
			url := fmt.Sprintf("%s:%d/update/gauge/%s/%.2f", host, port, metrics.Label, metrics.Value)
			urls[index] = url
		}

		PollCount++
		url := fmt.Sprintf("%s:%d/update/counter/PollCount/%d", host, port, PollCount)
		urls[urlsLength-2] = url

		url = fmt.Sprintf("%s:%d/update/gauge/RandomValue/%2.f", host, port, rand.Float64()*100)
		urls[urlsLength-1] = url

		toSend = append(toSend, urls...)

		time.Sleep(pollInterval)
		sleepBeforeSend += uint64(pollInterval)

		if sleepBeforeSend == uint64(reportInterval) {
			sleepBeforeSend = 0

			for _, url = range toSend {
				request, err := http.NewRequest(http.MethodPost, url, nil)
				if err != nil {
					panic(err)
				}

				resp, err := client.Do(request)
				if err != nil {
					panic(err)
				}
				defer resp.Body.Close()
			}

			toSend = slices.Delete(toSend, 0, len(toSend))
			toSend = toSend[:0]
		}
	}
}

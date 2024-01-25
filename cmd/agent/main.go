package main

import (
	"time"
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/metricspoll"
)

func runPoll(interval time.Duration, mh metricspoll.MetricsHandler) {
	for {
		mh.RefreshData()
		time.Sleep(interval)
	}
}

func runReport(interval time.Duration, mh metricspoll.MetricsHandler) {
	for {
		mh.SendData()
		time.Sleep(interval)
	}
}

func forever() {
	for {
		time.Sleep(10 * time.Second)
	}
}

func main() {

	args := config.ClientConfig{}
	args.Load()

	mh := metricspoll.NewMetricsHandler(args)
	mh.RefreshData()

	go runPoll(args.PollInterval, mh)
	go runReport(args.ReportInterval, mh)

	forever()

}

package main

import (
	"flag"
	"time"
	"yaprakticum-go-track2/internal/metricsPoll"
)

func main() {

	endp := flag.String("a", "localhost:8080", "Server endpoint address:port")
	pollInterval := flag.Float64("p", 2, "pollInterval")
	reportInterval := flag.Float64("r", 10, "reportInterval")
	flag.Parse()

	pollInterval_ := time.Duration(*pollInterval) * time.Second
	reportInterval_ := time.Duration(*reportInterval) * time.Second

	mh := metricsPoll.NewMetricsHandler(*endp)

	lastPoll := time.Now()
	lastReport := time.Now()

	for {

		time.Sleep(50 * time.Millisecond)

		if (time.Now().Sub(lastPoll)) >= pollInterval_ {
			lastPoll = time.Now()
			mh.RefreshData()
		}

		if (time.Now().Sub(lastReport)) >= reportInterval_ {
			lastReport = time.Now()
			mh.SendData()
		}
	}
}

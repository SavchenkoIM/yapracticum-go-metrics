package main

import (
	"flag"
	"os"
	"strconv"
	"time"
	"yaprakticum-go-track2/internal/metricspoll"
)

func main() {

	endp := flag.String("a", "localhost:8080", "Server endpoint address:port")
	pollInterval := flag.Float64("p", 2, "pollInterval")
	reportInterval := flag.Float64("r", 10, "reportInterval")
	flag.Parse()

	if val, exist := os.LookupEnv("ADDRESS"); exist {
		*endp = val
	}

	if _, exist := os.LookupEnv("REPORT_INTERVAL"); exist {
		if val, err := strconv.ParseFloat(os.Getenv("REPORT_INTERVAL"), 64); err != nil {
			*reportInterval = val
		}
	}
	if _, exist := os.LookupEnv("POLL_INTERVAL"); exist {
		if val, err := strconv.ParseFloat(os.Getenv("POLL_INTERVAL"), 64); err != nil {
			*pollInterval = val
		}
	}

	pollInterval_ := time.Duration(*pollInterval) * time.Second
	reportInterval_ := time.Duration(*reportInterval) * time.Second

	mh := metricspoll.NewMetricsHandler(*endp)
	mh.RefreshData()

	lastPoll := time.Now()
	lastReport := time.Now()

	for {

		time.Sleep(50 * time.Millisecond)

		if (time.Since(lastPoll)) >= pollInterval_ {
			lastPoll = time.Now()
			mh.RefreshData()
		}

		if (time.Since(lastReport)) >= reportInterval_ {
			lastReport = time.Now()
			mh.SendData()
		}
	}
}

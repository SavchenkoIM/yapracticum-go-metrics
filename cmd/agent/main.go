package main

import (
	"flag"
	"os"
	"strconv"
	"time"
	"yaprakticum-go-track2/internal/metricspoll"
)

type cliEnvArgs struct {
	endp           string
	pollInterval   time.Duration
	reportInterval time.Duration
}

func getCliEnvArgs() cliEnvArgs {
	var res cliEnvArgs
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

	res.endp = *endp
	res.pollInterval = time.Duration(*pollInterval) * time.Second
	res.reportInterval = time.Duration(*reportInterval) * time.Second
	return res
}

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

	args := getCliEnvArgs()

	mh := metricspoll.NewMetricsHandler(args.endp)
	mh.RefreshData()

	go runPoll(args.pollInterval, mh)
	go runReport(args.reportInterval, mh)

	forever()

}

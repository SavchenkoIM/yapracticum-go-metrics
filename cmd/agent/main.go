package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/metricspoll"
)

type metricsHandlerFunc func(context.Context)

func worker(mhf metricsHandlerFunc, name string, wg *sync.WaitGroup, ctx context.Context, interval time.Duration) {
	defer wg.Done()

	tck := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Stopping %s routine...\n", name)
			return
		case <-tck.C:
			mhf(ctx)
		}
	}
}

func main() {

	args := config.ClientConfig{}
	args.Load()

	parentContext := context.Background()
	cWithCancel, cancel := context.WithCancel(parentContext)

	mh := metricspoll.NewMetricsHandler(args)
	mh.RefreshDataWithSend(cWithCancel, false)

	wg := sync.WaitGroup{}

	wg.Add(3)
	go worker(mh.RefreshData, "poll", &wg, cWithCancel, args.PollInterval)
	go worker(mh.RefreshDataExt, "poll ext", &wg, cWithCancel, args.PollInterval)
	go worker(mh.SendData, "send", &wg, cWithCancel, args.PollInterval)

	go catchSignals(cancel)

	wg.Wait()
}

func catchSignals(cancel context.CancelFunc) {
	terminateSignals := make(chan os.Signal, 1)
	signal.Notify(terminateSignals, syscall.SIGINT, syscall.SIGTERM)
	<-terminateSignals
	cancel()
}

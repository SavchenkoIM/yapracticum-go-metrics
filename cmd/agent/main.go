<<<<<<< HEAD
// Agent of "metrics and alerting collecting system"

package main

import (
	"context"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/metricspoll"
	"yaprakticum-go-track2/internal/shared"
)

// Metrics handler function prototype (see package metricspoll)
type metricshandlerFunc func(context.Context)

// Routine for automatic refreshing of runtime metrics data
func agentRefreshRoutine(ctx context.Context, mhf metricshandlerFunc, name string, wg *sync.WaitGroup, interval time.Duration) {
	defer wg.Done()

	tck := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			shared.Logger.Sugar().Infof("Stopping %s routine...", name)
			return
		default:
			<-tck.C
			mhf(ctx)
		}
	}
}

// Routine worker for sending collected data to server of the system
//
// Those workers are started by agentSendRoutine function
func agentSendWorker(ctx context.Context, mhf metricshandlerFunc, actChan <-chan time.Time, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	shared.Logger.Info("Started report worker")

	for {
		select {
		case <-ctx.Done():
			shared.Logger.Info("Stopping report worker...")
			return
		case <-actChan:
			// Using "default: <-actChan" cause infinite lock of for loop for not started workers.
			// We'll chat in during 1-1 zoom session
			mhf(ctx)
		}
	}
}

// Starts maxWorkers agentSendWorker workers
func agentSendRoutine(ctx context.Context, mhf metricshandlerFunc, maxWorkers int, wg *sync.WaitGroup, interval time.Duration) {
	defer wg.Done()

	wChan := make(chan time.Time, 1000)
	for i := 0; i < maxWorkers; i++ {
		go agentSendWorker(ctx, mhf, wChan, wg)
	}

	tck := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			shared.Logger.Info("Stopping reporter routine...")
			return
		default:
			wChan <- <-tck.C
		}
	}
}

// Entry point of Agent
func main() {

	args := config.ClientConfig{}
	args.Load()

	var err error
	shared.Logger, err = zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	parentContext := context.Background()
	cWithCancel, cancel := context.WithCancel(parentContext)

	mh := metricspoll.NewMetricsHandler(args)
	mh.RefreshData(cWithCancel)

	wg := sync.WaitGroup{}

	wg.Add(2) // Add 2 not in routine to enshure that wg.Wait will work
	go agentRefreshRoutine(cWithCancel, mh.RefreshData, "main metrics", &wg, args.PollInterval)
	go agentRefreshRoutine(cWithCancel, mh.RefreshDataExt, "extended metrics", &wg, args.PollInterval)
	wg.Add(1) // additionally wg.Add(1) placed into each worker
	go agentSendRoutine(cWithCancel, mh.SendData, int(args.ReqLimit), &wg, args.ReportInterval)

	go catchSignals(cancel)

	wg.Wait()
}

// Handler of app termination signals
func catchSignals(cancel context.CancelFunc) {
	terminateSignals := make(chan os.Signal, 1)
	signal.Notify(terminateSignals, syscall.SIGINT, syscall.SIGTERM)
	<-terminateSignals
	cancel()
}
=======
package main

func main() {}
>>>>>>> template/up_to_iter24

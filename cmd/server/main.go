// Server of "metrics and alerting collecting system"

package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/handlers"
	"yaprakticum-go-track2/internal/shared"
	"yaprakticum-go-track2/internal/storage"
)

// Version info (are to be set by flags of go build)
var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

// Routine for periodic dump of metrics data to energy independed storage
func DumpDBFile(ctx context.Context, args config.ServerConfig, dataStorage *storage.Storage, logger *zap.Logger) {
	dt := time.Now()
	for {
		if args.StoreInterval > 0 {
			if time.Since(dt) >= args.StoreInterval {
				dt = time.Now()
				err := dataStorage.Dump(ctx)
				if err != nil {
					logger.Info(err.Error())
				}
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// Entry point of Server
func main() {
	fmt.Printf("Metrics Server\nBuild version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)

	cfg := config.ServerConfig{}
	args := cfg.Load()
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	var parentContext context.Context
	dataStorage, err := storage.InitStorage(parentContext, args, logger)
	if err != nil {
		panic(err)
	}
	defer dataStorage.Close(parentContext)
	handlers.SetDataStorage(dataStorage)
	handlers.SetConfig(cfg)

	handlers.SetDataStorage(dataStorage)

	shared.Logger = logger

	go DumpDBFile(parentContext, args, dataStorage, logger)

	server := http.Server{Addr: args.Endp, Handler: handlers.Router()}

	go catchSignal(parentContext, &server, dataStorage, logger)

	logger.Info("Server running at " + args.Endp)
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

// Handler of app termination signals
func catchSignal(ctx context.Context, server *http.Server, dataStorage *storage.Storage, logger *zap.Logger) {

	terminateSignals := make(chan os.Signal, 1)
	signal.Notify(terminateSignals, syscall.SIGINT, syscall.SIGTERM)

	//for {
	s := <-terminateSignals
	logger.Info("Got one of stop signals, shutting down server gracefully, SIGNAL NAME :" + s.String())

	dataStorage.Dump(ctx)
	server.Shutdown(ctx)
	//}
}

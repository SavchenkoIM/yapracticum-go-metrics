// Server of "metrics and alerting collecting system"

package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	hpprof "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/handlers"
	"yaprakticum-go-track2/internal/handlers/middleware"
	"yaprakticum-go-track2/internal/shared"
	"yaprakticum-go-track2/internal/storage"
)

// Router of the Server
func Router() chi.Router {

	r := chi.NewRouter()
	r.Use(middleware.GzipHandler, middleware.WithLogging)
	r.Route("/", func(r chi.Router) {
		r.Get("/", handlers.GetAllMetricsHandler)
		r.Route("/updates", func(r chi.Router) {
			r.Post("/", handlers.MultiMetricsUpdateHandlerREST)
		})
		r.Route("/update", func(r chi.Router) {
			r.Post("/", handlers.MetricsUpdateHandlerREST)
			r.Post("/{type}", func(res http.ResponseWriter, req *http.Request) {
				http.Error(res, "Not enough args (No name)", http.StatusNotFound)
			})
			r.Post("/{type}/{name}", func(res http.ResponseWriter, req *http.Request) {
				http.Error(res, "Not enough args (No value)", http.StatusBadRequest)
			})
			r.Post("/{type}/{name}/{value}", handlers.MetricUpdateHandler)

		})
		r.Route("/value", func(r chi.Router) {
			r.Get("/{type}/{name}", handlers.GetMetricHandler)
			r.Post("/", handlers.GetMetricHandlerREST)
		})
		r.Route("/ping", func(r chi.Router) {
			r.Get("/", handlers.PingHandler)
		})
		r.Route("/debug", func(r chi.Router) {
			r.Route("/pprof", func(r chi.Router) {
				r.Get("/", hpprof.Index)
				r.Get("/heap", hpprof.Index)
				r.Get("/profile", hpprof.Profile)
			})
		})
	})
	return r

}

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

	server := http.Server{Addr: args.Endp, Handler: Router()}

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

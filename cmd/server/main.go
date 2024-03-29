package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/handlers/getmetrics"
	"yaprakticum-go-track2/internal/handlers/middleware"
	"yaprakticum-go-track2/internal/handlers/updatemetrics"
	"yaprakticum-go-track2/internal/shared"
	"yaprakticum-go-track2/internal/storage"
)

var dataStorage *storage.Storage

func Router() chi.Router {

	r := chi.NewRouter()
	r.Use(middleware.GzipHandler, middleware.WithLogging)
	r.Route("/", func(r chi.Router) {
		r.Get("/", getmetric.GetAllMetricsHandler)
		r.Route("/updates", func(r chi.Router) {
			r.Post("/", updatemetrics.MultiMetricsUpdateHandlerREST)
		})
		r.Route("/update", func(r chi.Router) {
			r.Post("/", updatemetrics.MetricsUpdateHandlerREST)
			r.Post("/{type}", func(res http.ResponseWriter, req *http.Request) {
				http.Error(res, "Not enough args (No name)", http.StatusNotFound)
			})
			r.Post("/{type}/{name}", func(res http.ResponseWriter, req *http.Request) {
				http.Error(res, "Not enough args (No value)", http.StatusBadRequest)
			})
			r.Post("/{type}/{name}/{value}", updatemetrics.MetricUpdateHandler)

		})
		r.Route("/value", func(r chi.Router) {
			r.Get("/{type}/{name}", getmetric.GetMetricHandler)
			r.Post("/", getmetric.GetMetricHandlerREST)
		})
		r.Route("/ping", func(r chi.Router) {
			r.Get("/", getmetric.PingHandler)
		})
	})
	return r

}

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

func main() {

	cfg := config.ServerConfig{}
	args := cfg.Load()
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	var parentContext context.Context
	dataStorage, err = storage.InitStorage(parentContext, args, logger)
	if err != nil {
		panic(err)
	}
	defer dataStorage.Close(parentContext)
	updatemetrics.SetDataStorage(dataStorage)
	updatemetrics.SetCongig(cfg)

	getmetric.SetDataStorage(dataStorage)

	shared.Logger = logger

	go DumpDBFile(parentContext, args, dataStorage, logger)

	server := http.Server{Addr: args.Endp, Handler: Router()}

	go catchSignal(parentContext, &server, logger)

	logger.Info("Server running at " + args.Endp)
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

func catchSignal(ctx context.Context, server *http.Server, logger *zap.Logger) {

	terminateSignals := make(chan os.Signal, 1)
	signal.Notify(terminateSignals, syscall.SIGINT, syscall.SIGTERM)

	//for {
	s := <-terminateSignals
	logger.Info("Got one of stop signals, shutting down server gracefully, SIGNAL NAME :" + s.String())

	dataStorage.Dump(ctx)
	server.Shutdown(ctx)
	//}

}

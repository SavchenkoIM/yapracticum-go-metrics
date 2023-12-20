package main

import (
	"flag"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"yaprakticum-go-track2/internal/handlers/getmetrics"
	"yaprakticum-go-track2/internal/handlers/middleware"
	"yaprakticum-go-track2/internal/handlers/updatemetrics"
	"yaprakticum-go-track2/internal/storage"

	"github.com/go-chi/chi/v5"
)

var dataStorage storage.MemStorage

func Router() chi.Router {

	r := chi.NewRouter()
	r.Use(middleware.WithLogging)
	r.Route("/", func(r chi.Router) {
		r.Get("/", getmetric.GetAllMetricsHandler)
		r.Route("/update", func(r chi.Router) {
			r.Post("/", func(res http.ResponseWriter, req *http.Request) {
				http.Error(res, "Not enough args (No type)", http.StatusBadRequest)
			})
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
		})
	})
	return r

}

type srvEnvArgs struct {
	endp string
}

func getSrvEnvArgs() srvEnvArgs {
	var res srvEnvArgs
	endp := flag.String("a", ":8080", "Server endpoint address:port")

	if val, exist := os.LookupEnv("ADDRESS"); exist {
		*endp = val
	} else {
		flag.Parse()
	}

	res.endp = *endp
	return res
}

func main() {

	dataStorage = storage.InitStorage()
	updatemetrics.SetDataStorage(&dataStorage)
	getmetric.SetDataStorage(&dataStorage)
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	middleware.SetLogger(logger)

	args := getSrvEnvArgs()

	log.Println("Server running at " + args.endp)
	if err := http.ListenAndServe(args.endp, Router()); err != nil {
		panic(err)
	}
}

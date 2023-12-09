package main

import (
	"flag"
	"net/http"
	"os"
	"yaprakticum-go-track2/internal/handlers/getmetrics"
	"yaprakticum-go-track2/internal/handlers/middlware"
	"yaprakticum-go-track2/internal/handlers/updatemetrics"
	"yaprakticum-go-track2/internal/storage"

	"github.com/go-chi/chi/v5"
)

var dataStorage storage.MemStorage

func Router() chi.Router {

	mux := chi.NewRouter()
	mux.Use(middlware.CkeckIfAllCorrect)
	mux.Post("/update/{type}/{name}/{value}", updatemetrics.MetricUpdateHandler)
	mux.Get("/value/{type}/{name}", getmetric.GetMetricHandler)
	mux.Get("/", getmetric.GetAllMetricsHandler)
	return mux

}

func main() {

	endp := flag.String("a", ":8080", "Server endpoint address:port")

	if val, exist := os.LookupEnv("ADDRESS"); exist {
		*endp = val
	} else {
		flag.Parse()
	}

	dataStorage = storage.InitStorage()
	updatemetrics.SetDataStorage(&dataStorage)
	getmetric.SetDataStorage(&dataStorage)

	println("Server running at " + *endp)
	if err := http.ListenAndServe(*endp, Router()); err != nil {
		panic(err)
	}
}

/*mux.Post("/update/{type}/{name}/{value}", func(res http.ResponseWriter, req *http.Request) {
	typ := chi.URLParam(req, "type")
	println(typ)
})*/

//mux.Get("/update/", func(res http.ResponseWriter, req *http.Request) {
//	http.Error(res, "Server serves only POST requests", http.StatusBadRequest)
//})

//mux := http.NewServeMux()
//mux.Handle("/update/", http.StripPrefix("/update/", http.HandlerFunc(updateMetrics.MetricUpdateHandler)))

/*mux := chi.NewRouter()
mux.Use(middlware.CkeckIfAllCorrect)
mux.Post("/update/{type}/{name}/{value}", updateMetrics.MetricUpdateHandler)
mux.Get("/value/{type}/{name}", getMetric.GetMetricHandler)
mux.Get("/", getMetric.GetAllMetricsHandler)*/

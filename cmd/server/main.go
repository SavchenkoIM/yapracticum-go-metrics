package main

import (
	"handlers/updateMetrics"
	"net/http"
	"storage"
)

var dataStorage storage.MemStorage

func main() {

	dataStorage = storage.InitStorage()
	updateMetrics.SetDataStorage(&dataStorage)

	mux := http.NewServeMux()
	mux.Handle("/update/", http.StripPrefix("/update/", http.HandlerFunc(updateMetrics.MetricUpdateHandler)))

	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}

package getmetric

import (
	"fmt"
	"net/http"
	"strconv"
	"yaprakticum-go-track2/internal/storage"

	"github.com/go-chi/chi/v5"
)

var dataStorage *storage.MemStorage

func SetDataStorage(storage *storage.MemStorage) {
	dataStorage = storage
}

func GetAllMetricsHandler(res http.ResponseWriter, req *http.Request) {

	res.Write([]byte("=========================\n"))
	res.Write([]byte("COUNTERS:\n"))

	for k, v := range dataStorage.Counters.ReadData() {
		res.Write([]byte(fmt.Sprintf("%s: %d\n", k, v)))
	}

	res.Write([]byte("=========================\n"))
	res.Write([]byte("GAUGES:\n"))

	for k, v := range dataStorage.Gauges.ReadData() {
		res.Write([]byte(fmt.Sprintf("%s: %f\n", k, v)))
	}
}

func GetMetricHandler(res http.ResponseWriter, req *http.Request) {

	typ := chi.URLParam(req, "type")
	nam := chi.URLParam(req, "name")

	value, err := dataStorage.ReadData(typ, nam)

	if err != nil {
		http.Error(res, err.Error(), http.StatusNotFound)
		return
	}

	if typ == "counter" {
		res.Write([]byte(fmt.Sprintf("%d", value)))
	} else {
		res.Write([]byte(strconv.FormatFloat(value.(float64), 'f', -1, 64)))
	}

}

package getmetric

import (
	"encoding/json"
	"fmt"
	"log"
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

	dta1, _ := dataStorage.Counters.ReadData()
	for k, v := range dta1 {
		res.Write([]byte(fmt.Sprintf("%s: %d\n", k, v)))
	}

	res.Write([]byte("=========================\n"))
	res.Write([]byte("GAUGES:\n"))

	dta2, _ := dataStorage.Gauges.ReadData()
	for k, v := range dta2 {
		res.Write([]byte(fmt.Sprintf("%s: %f\n", k, v)))
	}
}

func GetMetricHandler(res http.ResponseWriter, req *http.Request) {

	typ := chi.URLParam(req, "type")
	nam := chi.URLParam(req, "name")

	switch typ {
	case "gauge":
		value, err := dataStorage.Gauges.ReadData(nam)
		if err != nil {
			http.Error(res, err.Error(), http.StatusNotFound)
			return
		}
		res.Write([]byte(strconv.FormatFloat(value[nam], 'f', -1, 64)))
	case "counter":
		value, err := dataStorage.Counters.ReadData(nam)
		if err != nil {
			http.Error(res, err.Error(), http.StatusNotFound)
			return
		}
		res.Write([]byte(fmt.Sprintf("%d", value[nam])))
	default:
		http.Error(res, "Unknown type "+typ, http.StatusNotFound)
		return
	}

}

func GetMetricHandlerREST(res http.ResponseWriter, req *http.Request) {

	var dta storage.Metrics

	body := make([]byte, req.ContentLength)
	req.Body.Read(body)
	req.Body.Close()

	log.Println("GetMetrics: " + string(body))

	err := json.Unmarshal(body, &dta)
	if err != nil {
		http.Error(res, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	dta2, err := dataStorage.ReadData(dta)

	if err == nil {
		resp, _ := json.MarshalIndent(dta2, "", "    ")
		res.Header().Set("Content-Type", "application/json")
		res.Write(resp)
		return
	}

	http.Error(res, err.Error(), http.StatusNotFound)

}

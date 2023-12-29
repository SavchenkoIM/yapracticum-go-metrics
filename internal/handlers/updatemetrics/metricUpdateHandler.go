package updatemetrics

import (
	"encoding/json"
	"net/http"
	"yaprakticum-go-track2/internal/storage"

	"github.com/go-chi/chi/v5"
)

var dataStorage *storage.MemStorage

func SetDataStorage(storage *storage.MemStorage) {
	dataStorage = storage
}

func MetricUpdateHandler(res http.ResponseWriter, req *http.Request) {

	typ := chi.URLParam(req, "type")
	name := chi.URLParam(req, "name")
	val := chi.URLParam(req, "value")

	switch typ {
	case "gauge":

		if err := dataStorage.Gauges.WriteData(name, val); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

	case "counter":

		if err := dataStorage.Counters.WriteData(name, val); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

	default:

		http.Error(res, "Unknown metric type: "+typ, http.StatusBadRequest)
		return
	}

}

func MetricsUpdateHandlerREST(res http.ResponseWriter, req *http.Request) {
	var dta storage.Metrics

	body := make([]byte, req.ContentLength)
	req.Body.Read(body)
	req.Body.Close()

	err := json.Unmarshal(body, &dta)
	if err != nil {
		http.Error(res, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	resp, err := dataStorage.WriteData(dta)
	val, _ := json.MarshalIndent(resp, "", "    ")

	if err == nil {
		res.Header().Set("Content-Type", "application/json")
		res.Write(val)
		return
	}

	http.Error(res, err.Error(), http.StatusBadRequest)
}

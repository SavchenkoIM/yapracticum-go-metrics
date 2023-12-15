package updateMetrics

import (
	"net/http"
	"storage"
	"strings"
)

var dataStorage *storage.MemStorage

func SetDataStorage(storage *storage.MemStorage) {
	dataStorage = storage
}

func MetricUpdateHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Server serves only POST requests", http.StatusBadRequest)
		return
	}

	pathWoSlash, _ := strings.CutSuffix(req.URL.Path, "/")
	pathWoSlash, _ = strings.CutPrefix(pathWoSlash, "/")
	reqData := strings.Split(pathWoSlash, "/")
	if len(reqData) < 2 {
		http.Error(res, "Not enough args (No name)", http.StatusNotFound)
		return
	}
	if len(reqData) < 3 {
		http.Error(res, "Not enough args (No type)", http.StatusBadRequest)
		return
	}

	switch reqData[0] {
	case "gauge":
		if err := dataStorage.Gauges.WriteData(reqData[1], reqData[2]); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
	case "counter":
		if err := dataStorage.Counters.WriteData(reqData[1], reqData[2]); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
	default:
		http.Error(res, "Unknown metric type", http.StatusBadRequest)
		return
	}
}

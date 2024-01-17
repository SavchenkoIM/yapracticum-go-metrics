package getmetric

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"yaprakticum-go-track2/internal/storage"

	"github.com/go-chi/chi/v5"
)

var dataStorage *storage.MemStorage

func SetDataStorage(storage *storage.MemStorage) {
	dataStorage = storage
}

func PingHandler(res http.ResponseWriter, req *http.Request) {
	if err := dataStorage.PingDB(); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
	}
	res.WriteHeader(http.StatusOK)
}

func GetAllMetricsHandler(res http.ResponseWriter, req *http.Request) {

	text := strings.Builder{}

	res.Header().Set("Content-Type", "text/html")

	text.WriteString("=========================\n")
	text.WriteString("COUNTERS:\n")

	dta1, _ := dataStorage.Counters.ReadData()
	for k, v := range dta1 {
		text.WriteString(fmt.Sprintf("%s: %d\n", k, v))
	}

	text.WriteString("=========================\n")
	text.WriteString("GAUGES:\n")

	dta2, _ := dataStorage.Gauges.ReadData()
	for k, v := range dta2 {
		text.WriteString(fmt.Sprintf("%s: %f\n", k, v))
	}

	txt := strings.Replace(text.String(), "\n", "</br>", -1)
	res.Write([]byte(txt))
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

	//log.Println("GetMetrics: " + string(body))

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

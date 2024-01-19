package getmetric

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"yaprakticum-go-track2/internal/storage"
	"yaprakticum-go-track2/internal/storage/storagecommons"

	"github.com/go-chi/chi/v5"
)

var dataStorage *storage.Storage

func SetDataStorage(storage *storage.Storage) {
	dataStorage = storage
}

func PingHandler(res http.ResponseWriter, req *http.Request) {
	if err := dataStorage.Ping(req.Context()); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
	}
	res.WriteHeader(http.StatusOK)
}

func GetAllMetricsHandler(res http.ResponseWriter, req *http.Request) {

	type Counter struct {
		Key   string
		Value int64
	}

	type Gauge struct {
		Key   string
		Value float64
	}

	type PageData struct {
		Counters []Counter
		Gauges   []Gauge
	}

	var pageData PageData
	pageData.Counters = make([]Counter, 0)
	pageData.Gauges = make([]Gauge, 0)

	dta1, _ := dataStorage.GetCounters().ReadData(req.Context())
	for k, v := range dta1 {
		pageData.Counters = append(pageData.Counters, Counter{k, v})
	}

	dta2, _ := dataStorage.GetGauges().ReadData(req.Context())
	for k, v := range dta2 {
		pageData.Gauges = append(pageData.Gauges, Gauge{k, v})
	}

	tmplStr := `=========================</br>
COUNTERS:</br>
{{range .Counters}}
	{{.Key}}:{{.Value}}</br>
{{end}}
=========================</br>
GAUGES:</br>
{{range .Gauges}} 
	{{.Key}}:{{.Value}}</br> 
{{end}}
`
	tmpl, _ := template.New("AllMetrics").Parse(tmplStr)

	res.Header().Set("Content-Type", "text/html")

	err := tmpl.Execute(res, pageData)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

}

func GetMetricHandler(res http.ResponseWriter, req *http.Request) {

	typ := chi.URLParam(req, "type")
	nam := chi.URLParam(req, "name")

	switch typ {
	case "gauge":
		value, err := dataStorage.GetGauges().ReadData(req.Context(), nam)
		if err != nil {
			http.Error(res, err.Error(), http.StatusNotFound)
			return
		}
		res.Write([]byte(strconv.FormatFloat(value[nam], 'f', -1, 64)))
	case "counter":
		value, err := dataStorage.GetCounters().ReadData(req.Context(), nam)
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

	var dta storagecommons.Metrics

	body := make([]byte, req.ContentLength)
	req.Body.Read(body)
	req.Body.Close()

	err := json.Unmarshal(body, &dta)
	if err != nil {
		http.Error(res, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	dta2, err := dataStorage.ReadData(req.Context(), dta)

	if err == nil {
		resp, _ := json.MarshalIndent(dta2, "", "    ")
		res.Header().Set("Content-Type", "application/json")
		res.Write(resp)
		//fmt.Println("GetMetrics: " + string(resp))
		return
	}

	http.Error(res, err.Error(), http.StatusNotFound)

}

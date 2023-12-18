package metricspoll

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
)

type metricsData struct {
	typ   string
	value string
}

type MetricsHandler struct {
	metricsMap map[string]metricsData
	counter    int64
	client     http.Client
}

func NewMetricsHandler(endp string) MetricsHandler {
	srvEndp = endp
	log.Println(srvEndp)
	return MetricsHandler{
		metricsMap: make(map[string]metricsData),
	}
}

func (ths *MetricsHandler) SendData() {
	for k, v := range ths.metricsMap {
		res, err := ths.client.Post("http://"+srvEndp+"/update/"+v.typ+"/"+k+"/"+v.value,
			"text/plain",
			nil)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		res.Body.Close()
	}
}

var srvEndp string

func (ths *MetricsHandler) RefreshData() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	ths.metricsMap["Alloc"] = metricsData{typ: "gauge", value: fmt.Sprintf("%v", ms.Alloc)}
	ths.metricsMap["BuckHashSys"] = metricsData{typ: "gauge", value: fmt.Sprintf("%v", ms.BuckHashSys)}
	ths.metricsMap["Frees"] = metricsData{typ: "gauge", value: fmt.Sprintf("%v", ms.Frees)}
	ths.metricsMap["GCCPUFraction"] = metricsData{typ: "gauge", value: fmt.Sprintf("%v", ms.GCCPUFraction)}
	ths.metricsMap["GCSys"] = metricsData{typ: "gauge", value: fmt.Sprintf("%v", ms.GCSys)}
	ths.metricsMap["HeapAlloc"] = metricsData{typ: "gauge", value: fmt.Sprintf("%v", ms.HeapAlloc)}
	ths.metricsMap["HeapIdle"] = metricsData{typ: "gauge", value: fmt.Sprintf("%v", ms.HeapIdle)}
	ths.metricsMap["HeapInuse"] = metricsData{typ: "gauge", value: fmt.Sprintf("%v", ms.HeapInuse)}
	ths.metricsMap["HeapObjects"] = metricsData{typ: "gauge", value: fmt.Sprintf("%v", ms.HeapObjects)}
	ths.metricsMap["HeapReleased"] = metricsData{typ: "gauge", value: fmt.Sprintf("%v", ms.HeapReleased)}
	ths.metricsMap["HeapSys"] = metricsData{typ: "gauge", value: fmt.Sprintf("%v", ms.HeapSys)}
	ths.metricsMap["LastGC"] = metricsData{typ: "gauge", value: fmt.Sprintf("%v", ms.LastGC)}
	ths.metricsMap["Lookups"] = metricsData{typ: "gauge", value: fmt.Sprintf("%v", ms.Lookups)}
	ths.metricsMap["MCacheInuse"] = metricsData{typ: "gauge", value: fmt.Sprintf("%v", ms.MCacheInuse)}
	ths.metricsMap["MCacheSys"] = metricsData{typ: "gauge", value: fmt.Sprintf("%v", ms.MCacheSys)}
	ths.metricsMap["MSpanInuse"] = metricsData{typ: "gauge", value: fmt.Sprintf("%v", ms.MSpanInuse)}
	ths.metricsMap["MSpanSys"] = metricsData{typ: "gauge", value: fmt.Sprintf("%v", ms.MSpanSys)}
	ths.metricsMap["Mallocs"] = metricsData{typ: "gauge", value: fmt.Sprintf("%v", ms.Mallocs)}
	ths.metricsMap["NextGC"] = metricsData{typ: "gauge", value: fmt.Sprintf("%v", ms.NextGC)}
	ths.metricsMap["NumForcedGC"] = metricsData{typ: "gauge", value: fmt.Sprintf("%v", ms.NumForcedGC)}
	ths.metricsMap["NumGC"] = metricsData{typ: "gauge", value: fmt.Sprintf("%v", ms.NumGC)}
	ths.metricsMap["OtherSys"] = metricsData{typ: "gauge", value: fmt.Sprintf("%v", ms.OtherSys)}
	ths.metricsMap["PauseTotalNs"] = metricsData{typ: "gauge", value: fmt.Sprintf("%v", ms.PauseTotalNs)}
	ths.metricsMap["StackInuse"] = metricsData{typ: "gauge", value: fmt.Sprintf("%v", ms.StackInuse)}
	ths.metricsMap["StackSys"] = metricsData{typ: "gauge", value: fmt.Sprintf("%v", ms.StackSys)}
	ths.metricsMap["Sys"] = metricsData{typ: "gauge", value: fmt.Sprintf("%v", ms.Sys)}
	ths.metricsMap["TotalAlloc"] = metricsData{typ: "gauge", value: fmt.Sprintf("%v", ms.TotalAlloc)}
	ths.metricsMap["RandomValue"] = metricsData{typ: "gauge", value: fmt.Sprintf("%v", rand.Float64())}
	//ths.metricsMap[28].value = fmt.Sprintf("%d", this.counter)

	res, err := ths.client.Post("http://"+srvEndp+"/update/counter/PollCount/1",
		"text/plain",
		nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer res.Body.Close()
}

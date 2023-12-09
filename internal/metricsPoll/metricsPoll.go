package metricspoll

import (
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
)

type metricsData struct {
	name  string
	typ   string
	value string
}

type MetricsHandler struct {
	metricsSlice []metricsData
	counter      int64
	client       http.Client
}

func NewMetricsHandler(endp string) MetricsHandler {
	srvEndp = endp
	print(srvEndp)
	return MetricsHandler{
		metricsSlice: []metricsData{
			{name: "Alloc", typ: "gauge", value: ""},
			{name: "BuckHashSys", typ: "gauge", value: ""},
			{name: "Frees", typ: "gauge", value: ""},
			{name: "GCCPUFraction", typ: "gauge", value: ""},
			{name: "GCSys", typ: "gauge", value: ""},
			{name: "HeapAlloc", typ: "gauge", value: ""},
			{name: "HeapIdle", typ: "gauge", value: ""},
			{name: "HeapInuse", typ: "gauge", value: ""},
			{name: "HeapObjects", typ: "gauge", value: ""},
			{name: "HeapReleased", typ: "gauge", value: ""},
			{name: "HeapSys", typ: "gauge", value: ""},
			{name: "LastGC", typ: "gauge", value: ""},
			{name: "Lookups", typ: "gauge", value: ""},
			{name: "MCacheInuse", typ: "gauge", value: ""},
			{name: "MCacheSys", typ: "gauge", value: ""},
			{name: "MSpanInuse", typ: "gauge", value: ""},
			{name: "MSpanSys", typ: "gauge", value: ""},
			{name: "Mallocs", typ: "gauge", value: ""},
			{name: "NextGC", typ: "gauge", value: ""},
			{name: "NumForcedGC", typ: "gauge", value: ""},
			{name: "NumGC", typ: "gauge", value: ""},
			{name: "OtherSys", typ: "gauge", value: ""},
			{name: "PauseTotalNs", typ: "gauge", value: ""},
			{name: "StackInuse", typ: "gauge", value: ""},
			{name: "StackSys", typ: "gauge", value: ""},
			{name: "Sys", typ: "gauge", value: ""},
			{name: "TotalAlloc", typ: "gauge", value: ""},
			{name: "RandomValue", typ: "gauge", value: ""},
			//{name: "PollCount", typ: "counter", value: ""},
		},
	}
}

func (ths MetricsHandler) SendData() {
	for _, v := range ths.metricsSlice {
		_, err := ths.client.Post("http://"+srvEndp+"/update/"+v.typ+"/"+v.name+"/"+v.value,
			"text/plain",
			nil)
		//defer res.Body.Close()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}

var srvEndp string

func (ths *MetricsHandler) RefreshData() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	ths.metricsSlice[0].value = fmt.Sprintf("%v", ms.Alloc)
	ths.metricsSlice[1].value = fmt.Sprintf("%v", ms.BuckHashSys)
	ths.metricsSlice[2].value = fmt.Sprintf("%v", ms.Frees)
	ths.metricsSlice[3].value = fmt.Sprintf("%v", ms.GCCPUFraction)
	ths.metricsSlice[4].value = fmt.Sprintf("%v", ms.GCSys)
	ths.metricsSlice[5].value = fmt.Sprintf("%v", ms.HeapAlloc)
	ths.metricsSlice[6].value = fmt.Sprintf("%v", ms.HeapIdle)
	ths.metricsSlice[7].value = fmt.Sprintf("%v", ms.HeapInuse)
	ths.metricsSlice[8].value = fmt.Sprintf("%v", ms.HeapObjects)
	ths.metricsSlice[9].value = fmt.Sprintf("%v", ms.HeapReleased)
	ths.metricsSlice[10].value = fmt.Sprintf("%v", ms.HeapSys)
	ths.metricsSlice[11].value = fmt.Sprintf("%v", ms.LastGC)
	ths.metricsSlice[12].value = fmt.Sprintf("%v", ms.Lookups)
	ths.metricsSlice[13].value = fmt.Sprintf("%v", ms.MCacheInuse)
	ths.metricsSlice[14].value = fmt.Sprintf("%v", ms.MCacheSys)
	ths.metricsSlice[15].value = fmt.Sprintf("%v", ms.MSpanInuse)
	ths.metricsSlice[16].value = fmt.Sprintf("%v", ms.MSpanSys)
	ths.metricsSlice[17].value = fmt.Sprintf("%v", ms.Mallocs)
	ths.metricsSlice[18].value = fmt.Sprintf("%v", ms.NextGC)
	ths.metricsSlice[19].value = fmt.Sprintf("%v", ms.NumForcedGC)
	ths.metricsSlice[20].value = fmt.Sprintf("%v", ms.NumGC)
	ths.metricsSlice[21].value = fmt.Sprintf("%v", ms.OtherSys)
	ths.metricsSlice[22].value = fmt.Sprintf("%v", ms.PauseTotalNs)
	ths.metricsSlice[23].value = fmt.Sprintf("%v", ms.StackInuse)
	ths.metricsSlice[24].value = fmt.Sprintf("%v", ms.StackSys)
	ths.metricsSlice[25].value = fmt.Sprintf("%v", ms.Sys)
	ths.metricsSlice[26].value = fmt.Sprintf("%v", ms.TotalAlloc)
	ths.metricsSlice[27].value = fmt.Sprintf("%v", rand.Float64())
	//ths.metricsSlice[28].value = fmt.Sprintf("%d", this.counter)

	_, err := ths.client.Post("http://"+srvEndp+"/update/counter/PollCount/1",
		"text/plain",
		nil)
	//defer res.Body.Close()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

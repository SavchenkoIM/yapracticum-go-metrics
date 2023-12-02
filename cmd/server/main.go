package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type metricFloat64 struct {
	data map[string]float64
}

func (this metricFloat64) writeData(key string, value string) error {

	v, err := strconv.ParseFloat(value, 64)

	if err == nil {
		this.data[key] = v
	}

	v2, exist := this.data[key]
	if exist {
		println("Value of " + key + " is " + fmt.Sprintf("%f", v2))
	}
	return err
}

func NewMetricFloat64() metricFloat64 {
	var v metricFloat64
	v.data = make(map[string]float64)
	return v
}

type metricInt64Slice struct {
	data map[string][]int64
}

func (this metricInt64Slice) writeData(key string, value string) error {
	v, err := strconv.ParseInt(value, 10, 64)

	if err == nil {
		if _, keyExist := this.data[key]; !keyExist {
			this.data[key] = make([]int64, 0)
			println("Created key: " + key)
		}
		this.data[key] = append(this.data[key], v)
		println("Slice " + key + " of length " + fmt.Sprintf("%d", len(this.data[key])))
	}
	return err
}

func NewMetricInt64Slice() metricInt64Slice {
	var v metricInt64Slice
	v.data = make(map[string][]int64)
	return v
}

type MemStorage struct {
	gauges   metricFloat64
	counters metricInt64Slice
}

var dataStorage MemStorage

func metricUpdateHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Server serves only POST requests", http.StatusBadRequest)
		return
	}

	pathWoSlash, _ := strings.CutSuffix(req.URL.Path, "/")
	reqData := strings.Split(pathWoSlash, "/")
	if len(reqData) < 2 {
		http.Error(res, "Not enough args (No type)", http.StatusNotFound)
		return
	}
	if len(reqData) < 3 {
		http.Error(res, "Not enough args (No name)", http.StatusBadRequest)
		return
	}

	switch reqData[0] {
	case "gauge":
		if err := dataStorage.gauges.writeData(reqData[1], reqData[2]); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
	case "counter":
		if err := dataStorage.counters.writeData(reqData[1], reqData[2]); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
	default:
		http.Error(res, "Unknown metric type", http.StatusBadRequest)
		return
	}
}

func main() {

	dataStorage.gauges = NewMetricFloat64()
	dataStorage.counters = NewMetricInt64Slice()

	mux := http.NewServeMux()
	mux.Handle("/update/", http.StripPrefix("/update/", http.HandlerFunc(metricUpdateHandler)))

	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}

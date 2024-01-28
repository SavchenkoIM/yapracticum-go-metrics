package metricspoll

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"golang.org/x/sync/semaphore"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/storage/storagecommons"
)

type RetryFunc func(r *http.Request) (*http.Response, error)

func RetryRequest(ctx context.Context, f RetryFunc, r *http.Request) (*http.Response, error) {
	var err error
	var res *http.Response

	waitTimes := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

	for i, timeWait := range waitTimes {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		res, err = f(r)
		if err == nil {
			return res, nil
		}

		time.Sleep(timeWait)

		fmt.Printf("Request retry #%d\n", i+1)
	}

	fmt.Printf("Request failed permanently!\n")
	return nil, err
}

type metricsData struct {
	typ      string
	value    float64
	ctrValue int64
}

type MetricsHandler struct {
	metricsMap map[string]metricsData
	counter    int64
	client     http.Client
	cfg        config.ClientConfig
	semaphore  *semaphore.Weighted
}

func NewMetricsHandler(cfg config.ClientConfig) MetricsHandler {
	srvEndp = cfg.Endp
	log.Println(srvEndp)
	return MetricsHandler{
		metricsMap: make(map[string]metricsData),
		cfg:        cfg,
		semaphore:  semaphore.NewWeighted(cfg.ReqLimit),
	}
}

func compressGzip(b []byte) ([]byte, error) {
	var bb bytes.Buffer
	gz := gzip.NewWriter(&bb)
	_, err := gz.Write(b)
	if err != nil {
		return nil, err
	}
	err = gz.Close()
	if err != nil {
		return nil, err
	}
	return bb.Bytes(), nil
}

func addHmacSha256(req *http.Request, body []byte, key string) {
	if key != "" {
		hmc := hmac.New(sha256.New, []byte(key))
		hmc.Write(body)
		req.Header.Set("HashSHA256", hex.EncodeToString(hmc.Sum(nil)))
	}
}

func (ths *MetricsHandler) SendData(ctx context.Context) {
	var dta storagecommons.MetricsDB

	for k, v := range ths.metricsMap {
		k, v := k, v
		var dta_ storagecommons.Metrics
		dta_.MType = v.typ
		dta_.ID = k
		switch dta_.MType {
		case "gauge":
			dta_.Value = &v.value
		case "counter":
			dta_.Delta = &v.ctrValue
		}

		dta.MetricsDB = append(dta.MetricsDB, dta_)
	}

	// Compress data
	jm, _ := json.Marshal(dta.MetricsDB)
	b, _ := compressGzip(jm)
	bb := bytes.NewBuffer(b)

	req, _ := http.NewRequest(http.MethodPost, "http://"+srvEndp+"/updates/", bb)
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	addHmacSha256(req, jm, ths.cfg.Key)

	err := ths.semaphore.Acquire(ctx, 1)
	if err != nil {
		return
	}
	//res, err := ths.client.Do(req)
	res, err := RetryRequest(ctx, ths.client.Do, req)

	ths.semaphore.Release(1)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	res.Body.Close()
}

var srvEndp string

func (ths *MetricsHandler) RefreshDataExt(ctx context.Context) {
	v, _ := mem.VirtualMemory()
	ths.metricsMap["TotalMemory"] = metricsData{typ: "gauge", value: float64(v.Total)}
	ths.metricsMap["FreeMemory"] = metricsData{typ: "gauge", value: float64(v.Free)}

	cu, err := cpu.PercentWithContext(ctx, 5*time.Second, true)

	if err == nil {
		for i, v := range cu {
			ths.metricsMap[fmt.Sprintf("CPUutilization%d", i+1)] = metricsData{typ: "gauge", value: v}
		}
	}

}

func (ths *MetricsHandler) RefreshData(ctx context.Context) {
	ths.RefreshDataWithSend(ctx, true)
}

func (ths *MetricsHandler) RefreshDataWithSend(ctx context.Context, sendCounter bool) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	ths.metricsMap["Alloc"] = metricsData{typ: "gauge", value: float64(ms.Alloc)}
	ths.metricsMap["BuckHashSys"] = metricsData{typ: "gauge", value: float64(ms.BuckHashSys)}
	ths.metricsMap["Frees"] = metricsData{typ: "gauge", value: float64(ms.Frees)}
	ths.metricsMap["GCCPUFraction"] = metricsData{typ: "gauge", value: float64(ms.GCCPUFraction)}
	ths.metricsMap["GCSys"] = metricsData{typ: "gauge", value: float64(ms.GCSys)}
	ths.metricsMap["HeapAlloc"] = metricsData{typ: "gauge", value: float64(ms.HeapAlloc)}
	ths.metricsMap["HeapIdle"] = metricsData{typ: "gauge", value: float64(ms.HeapIdle)}
	ths.metricsMap["HeapInuse"] = metricsData{typ: "gauge", value: float64(ms.HeapInuse)}
	ths.metricsMap["HeapObjects"] = metricsData{typ: "gauge", value: float64(ms.HeapObjects)}
	ths.metricsMap["HeapReleased"] = metricsData{typ: "gauge", value: float64(ms.HeapReleased)}
	ths.metricsMap["HeapSys"] = metricsData{typ: "gauge", value: float64(ms.HeapSys)}
	ths.metricsMap["LastGC"] = metricsData{typ: "gauge", value: float64(ms.LastGC)}
	ths.metricsMap["Lookups"] = metricsData{typ: "gauge", value: float64(ms.Lookups)}
	ths.metricsMap["MCacheInuse"] = metricsData{typ: "gauge", value: float64(ms.MCacheInuse)}
	ths.metricsMap["MCacheSys"] = metricsData{typ: "gauge", value: float64(ms.MCacheSys)}
	ths.metricsMap["MSpanInuse"] = metricsData{typ: "gauge", value: float64(ms.MSpanInuse)}
	ths.metricsMap["MSpanSys"] = metricsData{typ: "gauge", value: float64(ms.MSpanSys)}
	ths.metricsMap["Mallocs"] = metricsData{typ: "gauge", value: float64(ms.Mallocs)}
	ths.metricsMap["NextGC"] = metricsData{typ: "gauge", value: float64(ms.NextGC)}
	ths.metricsMap["NumForcedGC"] = metricsData{typ: "gauge", value: float64(ms.NumForcedGC)}
	ths.metricsMap["NumGC"] = metricsData{typ: "gauge", value: float64(ms.NumGC)}
	ths.metricsMap["OtherSys"] = metricsData{typ: "gauge", value: float64(ms.OtherSys)}
	ths.metricsMap["PauseTotalNs"] = metricsData{typ: "gauge", value: float64(ms.PauseTotalNs)}
	ths.metricsMap["StackInuse"] = metricsData{typ: "gauge", value: float64(ms.StackInuse)}
	ths.metricsMap["StackSys"] = metricsData{typ: "gauge", value: float64(ms.StackSys)}
	ths.metricsMap["Sys"] = metricsData{typ: "gauge", value: float64(ms.Sys)}
	ths.metricsMap["TotalAlloc"] = metricsData{typ: "gauge", value: float64(ms.TotalAlloc)}
	ths.metricsMap["RandomValue"] = metricsData{typ: "gauge", value: rand.Float64()}
	//ths.metricsMap[28].value = fmt.Sprintf("%d", this.counter)

	v := int64(1)
	jm, _ := json.Marshal(storagecommons.Metrics{MType: "counter", Delta: &v, ID: "PollCount"})
	b, _ := compressGzip(jm)
	bb := bytes.NewBuffer(b)

	req, _ := http.NewRequest(http.MethodPost, "http://"+srvEndp+"/update/", bb)
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	addHmacSha256(req, jm, ths.cfg.Key)

	//res, err := ths.client.Do(req)
	if sendCounter {
		err := ths.semaphore.Acquire(ctx, 1)
		if err != nil {
			return
		}

		res, err := RetryRequest(ctx, ths.client.Do, req)

		ths.semaphore.Release(1)

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		defer res.Body.Close()

	}

}

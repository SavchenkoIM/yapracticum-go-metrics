// Package contains tools for calculating and sending metrics data for Agent

package metricspoll

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"hash"
	"io"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/shared"
	"yaprakticum-go-track2/internal/storage/storagecommons"
)

// Counter of refresh data calls since last send
var accumPollCounter atomic.Int64

// Request function prototype for RetryRequest method
type RetryFunc func(r *http.Request) (*http.Response, error)

// Calls function f with retry in case of error returned
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

		if i < len(waitTimes)-1 {
			shared.Logger.Sugar().Infof("Request retry #%d", i+1)
		} else {
			shared.Logger.Sugar().Infof("Request failed permanently!")
		}
	}

	return nil, err
}

// Describes single metric
type metricsData struct {
	typ      string
	value    float64
	ctrValue int64
}

// Layer for complex metrics poll and send handling
type MetricsHandler struct {
	metricsMap      map[string]metricsData
	counter         int64
	client          http.Client
	cfg             config.ClientConfig
	metricsMapMutex sync.RWMutex
}

// Constructor for MetricsHandler
func NewMetricsHandler(cfg config.ClientConfig) MetricsHandler {
	srvEndp = cfg.Endp
	shared.Logger.Info(srvEndp)
	return MetricsHandler{
		metricsMap: make(map[string]metricsData),
		cfg:        cfg,
	}
}

// Auxilary function for compressing request body
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

// Auxilary function for adding HMAC signature to request
func addHmacSha256(req *http.Request, body []byte, key string) *http.Request {
	if key != "" {
		hmc := hmac.New(sha256.New, []byte(key))
		hmc.Write(body)
		req.Header.Set("HashSHA256", hex.EncodeToString(hmc.Sum(nil)))
	}
	return req
}

func EncryptOAEP(hash hash.Hash, random io.Reader, public *rsa.PublicKey, msg []byte, label []byte) ([]byte, error) {
	msgLen := len(msg)
	step := public.Size() - 2*hash.Size() - 2
	var encryptedBytes []byte

	for start := 0; start < msgLen; start += step {
		finish := start + step
		if finish > msgLen {
			finish = msgLen
		}

		encryptedBlockBytes, err := rsa.EncryptOAEP(hash, random, public, msg[start:finish], label)
		if err != nil {
			return nil, err
		}

		encryptedBytes = append(encryptedBytes, encryptedBlockBytes...)
	}

	return encryptedBytes, nil
}

func getEncryptedBody(body []byte, cfg config.ClientConfig) ([]byte, error) {
	if !cfg.UseRSA {
		return body, nil
	}

	oaep, err := EncryptOAEP(sha256.New(), crand.Reader, &cfg.RSAPublicKey, body, nil)
	if err != nil {
		return nil, err
	}

	return oaep, nil
}

// Filling JSON marshallable structure to be sent to Server
func (ths *MetricsHandler) prepareData() storagecommons.MetricsDB {
	var dta storagecommons.MetricsDB

	ths.metricsMapMutex.Lock()
	defer ths.metricsMapMutex.Unlock()
	// Can not use RLock, as CR#1 requested not to send PollCounter during each Poll in Poll function as
	// required in Increment 2.
	// So agent need to send accumulated PollCounter value in Send function, and this leads to
	// changing metricsMap["PollCount"] before each Send

	ths.metricsMap["PollCount"] = metricsData{typ: "counter", ctrValue: accumPollCounter.Swap(0)}
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

	return dta
}

// Sends data to server
func (ths *MetricsHandler) SendData(ctx context.Context) {

	dta := ths.prepareData()

	// Compress data
	jm, _ := json.Marshal(dta.MetricsDB)
	b, _ := compressGzip(jm)
	b, err := getEncryptedBody(b, ths.cfg)

	if err != nil {
		shared.Logger.Sugar().Errorf("Error while prepering data: %v", err)
		return
	}

	bb := bytes.NewBuffer(b)

	req, _ := http.NewRequest(http.MethodPost, "http://"+srvEndp+"/updates/", bb)
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	if ths.cfg.RealIP != nil {
		req.Header.Set("X-Real-IP", ths.cfg.RealIP.String())
	}

	addHmacSha256(req, jm, ths.cfg.Key)

	//res, err := ths.client.Do(req)
	res, err := RetryRequest(ctx, ths.client.Do, req)

	if err != nil {
		shared.Logger.Sugar().Infof("Error: %v", err)
		return
	}
	res.Body.Close()
}

var srvEndp string

// Calculates extended metrics data
func (ths *MetricsHandler) RefreshDataExt(ctx context.Context) {
	v, _ := mem.VirtualMemory()
	cu, err := cpu.PercentWithContext(ctx, 5*time.Second, true)

	ths.metricsMapMutex.Lock()
	defer ths.metricsMapMutex.Unlock()

	ths.metricsMap["TotalMemory"] = metricsData{typ: "gauge", value: float64(v.Total)}
	ths.metricsMap["FreeMemory"] = metricsData{typ: "gauge", value: float64(v.Free)}

	if err == nil {
		for i, v := range cu {
			ths.metricsMap[fmt.Sprintf("CPUutilization%d", i+1)] = metricsData{typ: "gauge", value: v}
		}
	}

}

// Retrieves main metrics data
func (ths *MetricsHandler) RefreshData(ctx context.Context) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	ths.metricsMapMutex.Lock()
	defer ths.metricsMapMutex.Unlock()

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
	ths.metricsMap["RandomValue"] = metricsData{typ: "gauge", value: rand.Float64()}
	//ths.metricsMap["PollCount"] = metricsData{typ: "counter", ctrValue: 1}
	accumPollCounter.Add(1)

}

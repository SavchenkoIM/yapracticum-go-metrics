package storage

import (
	"encoding/json"
	"errors"
	"go.uber.org/zap"
	"os"
	"strconv"
	"sync"
	"yaprakticum-go-track2/internal/config"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type MetricsDB struct {
	MetricsDB []Metrics `json:"metrics_db"`
}

// Storage

type MemStorage struct {
	Gauges    *metricFloat64
	Counters  *metricInt64Sum
	dumpMutex sync.Mutex
	syncWrite bool
	fileName  string
}

func InitStorage(args config.ServerConfig, logger *zap.Logger) (*MemStorage, error) {
	var ms MemStorage
	ms.Counters = newMetricInt64Sum()
	ms.Gauges = newMetricFloat64()
	ms.syncWrite = args.StoreInterval == 0
	ms.fileName = args.FileStoragePath

	if args.Restore {
		err := ms.Load()
		if err != nil {
			logger.Sugar().Infof("Unable to load data from file: %s", err.Error())
		}
	}

	return &ms, nil
}

func (ms *MemStorage) Close() error {
	return ms.Dump()
}

func (ms *MemStorage) Dump() error {
	ms.dumpMutex.Lock()
	defer ms.dumpMutex.Unlock()

	mdb := MetricsDB{MetricsDB: make([]Metrics, 0)}
	for k, v := range ms.Gauges.data {
		v2 := v
		mdb.MetricsDB = append(mdb.MetricsDB, Metrics{
			ID:    k,
			MType: "gauge",
			Value: &v2,
		})
	}
	for k, v := range ms.Counters.data {
		v2 := v
		mdb.MetricsDB = append(mdb.MetricsDB, Metrics{
			ID:    k,
			MType: "counter",
			Delta: &v2,
		})
	}

	jsn, err := json.MarshalIndent(mdb, "", "    ")
	if err != nil {
		return err
	}

	f, err := os.OpenFile(ms.fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(jsn)
	if err != nil {
		return err
	}

	return nil
}

func (ms *MemStorage) Load() error {
	data, err := os.ReadFile(ms.fileName)
	if err != nil {
		return err
	}
	mdb := MetricsDB{MetricsDB: make([]Metrics, 0)}
	err = json.Unmarshal(data, &mdb)
	if err != nil {
		return err
	}

	for _, v := range mdb.MetricsDB {
		switch v.MType {
		case "counter":
			ms.Counters.WriteDataPP(v.ID, *v.Delta)
		case "gauge":
			ms.Gauges.WriteDataPP(v.ID, *v.Value)
		}
	}

	return nil
}

func (ms *MemStorage) WriteData(metrics Metrics) (rMetrics Metrics, rError error) {
	rError = nil
	rMetrics = metrics

	switch metrics.MType {
	case "gauge":
		if metrics.Value == nil {
			return metrics, errors.New("no Value data provided")
		}
		ms.Gauges.WriteDataPP(metrics.ID, *metrics.Value)
		rMetrics = metrics
	case "counter":
		if metrics.Delta == nil {
			return metrics, errors.New("no Value data provided")
		}
		ms.Counters.WriteDataPP(metrics.ID, *metrics.Delta)
		vl := ms.Counters.data[metrics.ID]
		metrics.Delta = &vl
		rMetrics = metrics
	default:
		rError = errors.New("Unknown metric type: " + metrics.MType)
	}

	if ms.syncWrite {
		ms.Dump()
	}

	return
}

func (ms *MemStorage) ReadData(metrics Metrics) (Metrics, error) {
	switch metrics.MType {
	case "gauge":
		vl, exist := ms.Gauges.data[metrics.ID]
		if exist {
			metrics.Value = &vl
			return metrics, nil
		}
		return metrics, errors.New("Key gauge/" + metrics.ID + " not exists")
	case "counter":
		vl, exist := ms.Counters.data[metrics.ID]
		if exist {
			metrics.Delta = &vl
			return metrics, nil
		}
		return metrics, errors.New("Key counter/" + metrics.ID + " not exists")
	default:
		return metrics, errors.New("Unknown metric type: " + metrics.MType)
	}
}

// Float64

type metricFloat64 struct {
	data map[string]float64
	mu   sync.Mutex
}

func newMetricFloat64() *metricFloat64 {
	var v metricFloat64
	v.data = make(map[string]float64)
	return &v
}

func (ths *metricFloat64) ReadData(keys ...string) (map[string]float64, error) {
	switch len(keys) {
	case 0:
		return ths.data, nil
	case 1:
		key := keys[0]
		val, exist := ths.data[key]
		if exist {
			return map[string]float64{key: val}, nil
		}
		return nil, errors.New("Key gauge/" + key + " not exists")
	default:
		return nil, errors.New("it is allowed to request only one key at a time")
	}
}

func (ths *metricFloat64) WriteData(key string, value string) error {
	v, err := strconv.ParseFloat(value, 64)

	if err == nil {
		ths.mu.Lock()
		ths.data[key] = v
		defer ths.mu.Unlock()
	}

	return err
}

func (ths *metricFloat64) WriteDataPP(key string, value float64) {
	ths.mu.Lock()
	ths.data[key] = value
	defer ths.mu.Unlock()
}

// Int64 Cumulative

type metricInt64Sum struct {
	data map[string]int64
	mu   sync.Mutex
}

func newMetricInt64Sum() *metricInt64Sum {
	var v metricInt64Sum
	v.data = make(map[string]int64)
	return &v
}

func (ths *metricInt64Sum) ReadData(keys ...string) (map[string]int64, error) {
	switch len(keys) {
	case 0:
		return ths.data, nil
	case 1:
		key := keys[0]
		val, exist := ths.data[key]
		if exist {
			return map[string]int64{key: val}, nil
		}
		return nil, errors.New("Key gauge/" + key + " not exists")
	default:
		return nil, errors.New("it is allowed to request only one key at a time")
	}
}

func (ths *metricInt64Sum) WriteData(key string, value string) error {

	v, err := strconv.ParseInt(value, 10, 64)

	if err == nil {
		ths.mu.Lock()
		ths.data[key] += v
		defer ths.mu.Unlock()
	}

	return err
}

func (ths *metricInt64Sum) WriteDataPP(key string, value int64) {
	ths.mu.Lock()
	ths.data[key] += value
	defer ths.mu.Unlock()
}

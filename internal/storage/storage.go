package storage

import (
	"errors"
	"strconv"
	"sync"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

// Storage

func InitStorage() MemStorage {
	var ms MemStorage
	ms.Counters = newMetricInt64Sum()
	ms.Gauges = newMetricFloat64()
	return ms
}

type MemStorage struct {
	Gauges   *metricFloat64
	Counters *metricInt64Sum
}

func (ms *MemStorage) WriteData(metrics Metrics) (Metrics, error) {
	switch metrics.MType {
	case "gauge":
		if metrics.Value == nil {
			return metrics, errors.New("no Value data provided")
		}
		ms.Gauges.WriteDataPP(metrics.ID, *metrics.Value)
		return metrics, nil
	case "counter":
		if metrics.Delta == nil {
			return metrics, errors.New("no Value data provided")
		}
		ms.Counters.WriteDataPP(metrics.ID, *metrics.Delta)
		vl := ms.Counters.data[metrics.ID]
		metrics.Delta = &vl
		return metrics, nil
	default:
		return metrics, errors.New("Unknown metric type: " + metrics.MType)
	}
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

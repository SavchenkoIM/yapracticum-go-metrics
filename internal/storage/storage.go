package storage

import (
	"errors"
	"strconv"
	"sync"
)

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
		return nil, errors.New("It is allowed to request only one key at a time")
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
		return nil, errors.New("It is allowed to request only one key at a time")
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

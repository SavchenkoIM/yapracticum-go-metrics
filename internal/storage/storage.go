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
	Gauges   metricFloat64
	Counters metricInt64Sum
}

func (ths MemStorage) ReadData(typ string, key string) (interface{}, error) {

	switch typ {
	case "counter":
		val, exist := ths.Counters.data[key]
		if exist {
			return val, nil
		}

		return nil, errors.New("Key counters/" + key + " not exists")

	case "gauge":
		val, exist := ths.Gauges.data[key]
		if exist {
			return val, nil
		}

		return nil, errors.New("Key gauge/" + key + " not exists")

	default:
		return nil, errors.New("Unknown type " + typ)
	}

}

// Float64

type metricFloat64 struct {
	data map[string]float64
	mu   sync.Mutex
}

func newMetricFloat64() metricFloat64 {
	var v metricFloat64
	v.data = make(map[string]float64)
	return v
}

func (ths metricFloat64) ReadData() map[string]float64 {
	return ths.data
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

func newMetricInt64Sum() metricInt64Sum {
	var v metricInt64Sum
	v.data = make(map[string]int64)
	return v
}

func (ths metricInt64Sum) ReadData() map[string]int64 {
	return ths.data
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

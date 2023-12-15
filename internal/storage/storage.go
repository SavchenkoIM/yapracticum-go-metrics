package storage

import (
	"errors"
	"fmt"
	"strconv"
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
		val, exist := ths.Counters.Data[key]
		if exist {
			return val, nil
		} else {
			return nil, errors.New("Key counters/" + key + " not exists")
		}
	case "gauge":
		val, exist := ths.Gauges.Data[key]
		if exist {
			return val, nil
		} else {
			return nil, errors.New("Key gauge/" + key + " not exists")
		}
	default:
		return nil, errors.New("Unknown type " + typ)
	}

}

// Float64

type metricFloat64 struct {
	Data map[string]float64
}

func newMetricFloat64() metricFloat64 {
	var v metricFloat64
	v.Data = make(map[string]float64)
	return v
}

func (ths metricFloat64) WriteData(key string, value string) error {
	v, err := strconv.ParseFloat(value, 64)

	if err == nil {
		ths.Data[key] = v
	}

	v2, exist := ths.Data[key]
	if exist {
		println("Value of " + key + " is " + fmt.Sprintf("%f", v2))
	}
	return err
}

// Int64 Cumulative

type metricInt64Sum struct {
	Data map[string]int64
}

func (ths metricInt64Sum) WriteData(key string, value string) error {

	v, err := strconv.ParseInt(value, 10, 64)

	if err == nil {
		ths.Data[key] += v
	}

	v2, exist := ths.Data[key]
	if exist {
		println("Value of " + key + " is " + fmt.Sprintf("%d", v2))
	}
	return err
}

func newMetricInt64Sum() metricInt64Sum {
	var v metricInt64Sum
	v.Data = make(map[string]int64)
	return v
}

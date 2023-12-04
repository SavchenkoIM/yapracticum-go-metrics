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

func (this MemStorage) ReadData(typ string, key string) (interface{}, error) {

	switch typ {
	case "counter":
		val, exist := this.Counters.data[key]
		if exist {
			return val, nil
		} else {
			return nil, errors.New("Key counters/" + key + " not exists")
		}
	case "gauge":
		val, exist := this.Gauges.data[key]
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
	data map[string]float64
}

func newMetricFloat64() metricFloat64 {
	var v metricFloat64
	v.data = make(map[string]float64)
	return v
}

func (this metricFloat64) WriteData(key string, value string) error {
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

// Int64 Cumulative

type metricInt64Sum struct {
	data map[string]int64
}

func (this metricInt64Sum) WriteData(key string, value string) error {

	v, err := strconv.ParseInt(value, 10, 64)

	if err == nil {
		this.data[key] += v
	}

	v2, exist := this.data[key]
	if exist {
		println("Value of " + key + " is " + fmt.Sprintf("%d", v2))
	}
	return err
}

func newMetricInt64Sum() metricInt64Sum {
	var v metricInt64Sum
	v.data = make(map[string]int64)
	return v
}

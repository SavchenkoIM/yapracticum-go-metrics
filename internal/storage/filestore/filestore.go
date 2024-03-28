package filestore

import (
	"context"
	"encoding/json"
	"errors"
	"go.uber.org/zap"
	"os"
	"strconv"
	"sync"
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/storage/storagecommons"
)

// InMemory, file-backed storage description (see Storager)
type FileStore struct {
	Gauges    *MetricFloat64
	Counters  *MetricInt64Sum
	dumpMutex sync.Mutex
	syncWrite bool
	fileName  string
}

func New(ctx context.Context, args config.ServerConfig, logger *zap.Logger) (*FileStore, error) {
	var ms FileStore

	logger.Sugar().Infof("Creating memory/file storage...")

	ms.syncWrite = args.StoreInterval == 0
	ms.fileName = args.FileStoragePath

	ms.Gauges = NewMetricFloat64()
	ms.Counters = NewMetricInt64Sum()

	if args.Restore {
		err := ms.Load(ctx)
		if err != nil {
			logger.Sugar().Infof("Unable to load data from file: %s", err.Error())
		}
	}

	return &ms, nil
}

// Float64

type MetricFloat64 struct {
	data map[string]float64
	mu   sync.Mutex
}

func NewMetricFloat64() *MetricFloat64 {
	return &MetricFloat64{data: make(map[string]float64)}
}

func (ths *MetricFloat64) ReadData(ctx context.Context, keys ...string) (map[string]float64, error) {
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

func (ths *MetricFloat64) WriteData(ctx context.Context, key string, value string) error {
	v, err := strconv.ParseFloat(value, 64)

	if err != nil {
		return err
	}

	ths.mu.Lock()
	ths.data[key] = v
	ths.mu.Unlock()

	return nil
}

func (ths *MetricFloat64) WriteDataPP(ctx context.Context, key string, value float64) error {
	ths.mu.Lock()
	ths.data[key] = value
	ths.mu.Unlock()
	return nil
}

// Int64 Cumulative

type MetricInt64Sum struct {
	data map[string]int64
	mu   sync.Mutex
}

func NewMetricInt64Sum() *MetricInt64Sum {
	return &MetricInt64Sum{data: make(map[string]int64)}
}

func (ths *MetricInt64Sum) ReadData(ctx context.Context, keys ...string) (map[string]int64, error) {
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

func (ths *MetricInt64Sum) WriteData(ctx context.Context, key string, value string) error {

	v, err := strconv.ParseInt(value, 10, 64)

	if err != nil {
		return err
	}

	ths.mu.Lock()
	ths.data[key] += v
	ths.mu.Unlock()
	return nil
}

func (ths *MetricInt64Sum) WriteDataPP(ctx context.Context, key string, value int64) error {
	ths.mu.Lock()
	ths.data[key] += value
	ths.mu.Unlock()

	return nil
}

// DumpLoad

func (ms *FileStore) Dump(ctx context.Context) error {
	ms.dumpMutex.Lock()
	defer ms.dumpMutex.Unlock()

	mdb := storagecommons.MetricsDB{MetricsDB: make([]storagecommons.Metrics, 0)}

	data, err := ms.Gauges.ReadData(ctx)
	if err != nil {
		return err
	}
	for k, v := range data {
		v2 := v
		mdb.MetricsDB = append(mdb.MetricsDB, storagecommons.Metrics{
			ID:    k,
			MType: "gauge",
			Value: &v2,
		})
	}

	data2, err := ms.Counters.ReadData(ctx)
	if err != nil {
		return err
	}
	for k, v := range data2 {
		v2 := v
		mdb.MetricsDB = append(mdb.MetricsDB, storagecommons.Metrics{
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

func (ms *FileStore) Load(ctx context.Context) error {
	data, err := os.ReadFile(ms.fileName)
	if err != nil {
		return err
	}
	mdb := storagecommons.MetricsDB{MetricsDB: make([]storagecommons.Metrics, 0)}
	err = json.Unmarshal(data, &mdb)
	if err != nil {
		return err
	}

	for _, v := range mdb.MetricsDB {
		switch v.MType {
		case "counter":
			ms.Counters.WriteDataPP(ctx, v.ID, *v.Delta)
		case "gauge":
			ms.Gauges.WriteDataPP(ctx, v.ID, *v.Value)
		}
	}

	return nil
}

func (ms *FileStore) WriteDataMulty(ctx context.Context, metrics storagecommons.MetricsDB) error {
	for _, record := range metrics.MetricsDB {
		_, err := ms.WriteData(ctx, record)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ms *FileStore) WriteData(ctx context.Context, metrics storagecommons.Metrics) (rMetrics storagecommons.Metrics, rError error) {
	rError = nil
	rMetrics = metrics

	switch metrics.MType {
	case "gauge":
		if metrics.Value == nil {
			return metrics, errors.New("no Value data provided")
		}
		ms.Gauges.WriteDataPP(ctx, metrics.ID, *metrics.Value)
		rMetrics = metrics
	case "counter":
		if metrics.Delta == nil {
			return metrics, errors.New("no Value data provided")
		}
		ms.Counters.WriteDataPP(ctx, metrics.ID, *metrics.Delta)

		data, err := ms.Counters.ReadData(ctx)
		if err != nil {
			return rMetrics, err
		}

		vl := data[metrics.ID]
		metrics.Delta = &vl
		rMetrics = metrics
	default:
		rError = errors.New("Unknown metric type: " + metrics.MType)
	}

	if ms.syncWrite {
		ms.Dump(ctx)
	}

	return
}

func (ms *FileStore) ReadData(ctx context.Context, metrics storagecommons.Metrics) (storagecommons.Metrics, error) {
	switch metrics.MType {
	case "gauge":
		data, err := ms.Gauges.ReadData(ctx)

		if err != nil {
			return metrics, err
		}

		vl, exist := data[metrics.ID]
		if exist {
			metrics.Value = &vl
			return metrics, nil
		}
		return metrics, errors.New("Key gauge/" + metrics.ID + " not exists")
	case "counter":
		data, err := ms.Counters.ReadData(ctx)

		if err != nil {
			return metrics, nil
		}

		vl, exist := data[metrics.ID]
		if exist {
			metrics.Delta = &vl
			return metrics, nil
		}
		return metrics, errors.New("Key counter/" + metrics.ID + " not exists")
	default:
		return metrics, errors.New("Unknown metric type: " + metrics.MType)
	}
}

func (ms *FileStore) Close(ctx context.Context) error {
	return nil
}

func (ms *FileStore) GetGauges() storagecommons.StoragerFloat64 {
	return ms.Gauges
}

func (ms *FileStore) GetCounters() storagecommons.StoragerInt64Sum {
	return ms.Counters
}

func (ms *FileStore) Ping(ctx context.Context) error {
	return nil
}

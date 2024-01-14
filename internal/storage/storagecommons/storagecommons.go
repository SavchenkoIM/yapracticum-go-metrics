package storagecommons

import "context"

type StoragerFloat64 interface {
	ReadData(ctx context.Context, keys ...string) (map[string]float64, error)
	WriteData(ctx context.Context, key string, value string) error
	WriteDataPP(ctx context.Context, key string, value float64) error
}

type StoragerInt64Sum interface {
	ReadData(ctx context.Context, keys ...string) (map[string]int64, error)
	WriteData(ctx context.Context, key string, value string) error
	WriteDataPP(ctx context.Context, key string, value int64) error
}

type Storager interface {
	Dump(ctx context.Context) error
	Load(ctx context.Context) error
	Close(ctx context.Context) error
	WriteDataMulty(ctx context.Context, metrics MetricsDB) error
	WriteData(ctx context.Context, metrics Metrics) (Metrics, error)
	ReadData(ctx context.Context, metrics Metrics) (Metrics, error)
	GetGauges() StoragerFloat64
	GetCounters() StoragerInt64Sum
	Ping(ctx context.Context) error
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type MetricsDB struct {
	MetricsDB []Metrics `json:"metrics_db"`
}

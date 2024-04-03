package storagecommons

import "context"

// Gauge storage interface
type StoragerFloat64 interface {
	// Read Gauge values for `keys` keys (if `keys` empty, returns all stored Gauges)
	ReadData(ctx context.Context, keys ...string) (map[string]float64, error)
	// Wrire Gauge data
	WriteData(ctx context.Context, key string, value string) error
	// Wrire Gauge data (value is pre-parsed)
	WriteDataPP(ctx context.Context, key string, value float64) error
}

// Counter storage interface
type StoragerInt64Sum interface {
	// Read Counter values for `keys` keys (if `keys` empty, returns all stored Counters)
	ReadData(ctx context.Context, keys ...string) (map[string]int64, error)
	// Wrire Counter data
	WriteData(ctx context.Context, key string, value string) error
	// Wrire Counter data (value is pre-parsed)
	WriteDataPP(ctx context.Context, key string, value int64) error
}

// Whole storage interface
type Storager interface {
	// Store data to power independed storage
	Dump(ctx context.Context) error
	// Load data from power independed storage
	Load(ctx context.Context) error
	// Destructor for Storage
	Close(ctx context.Context) error
	// Packet metrics write
	WriteDataMulty(ctx context.Context, metrics MetricsDB) error
	// Single metric write
	WriteData(ctx context.Context, metrics Metrics) (Metrics, error)
	// Returns JSON serializable structure with requested metric data
	ReadData(ctx context.Context, metrics Metrics) (Metrics, error)
	// Returns Gauges sub-storage object
	GetGauges() StoragerFloat64
	// Returns Counters sub-storage object
	GetCounters() StoragerInt64Sum
	// Check if storage is up
	Ping(ctx context.Context) error
}

// JSON serializable structure describing single metric
type Metrics struct {
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
}

// JSON serializable structure describing batch of metrics
type MetricsDB struct {
	MetricsDB []Metrics `json:"metrics_db"`
}

package storagecommons

type StoragerFloat64 interface {
	ReadData(keys ...string) (map[string]float64, error)
	WriteData(key string, value string) error
	WriteDataPP(key string, value float64) error
}

type StoragerInt64Sum interface {
	ReadData(keys ...string) (map[string]int64, error)
	WriteData(key string, value string) error
	WriteDataPP(key string, value int64) error
}

type Storager interface {
	Dump() error
	Load() error
	Close() error
	WriteDataMulty(metrics MetricsDB) error
	WriteData(metrics Metrics) (Metrics, error)
	ReadData(metrics Metrics) (Metrics, error)
	GetGauges() StoragerFloat64
	GetCounters() StoragerInt64Sum
	Ping() error
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

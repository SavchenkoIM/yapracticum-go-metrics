package dbstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"strconv"
	"sync"
	"time"
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/storage/storagecommons"
)

type DBStore struct {
	Gauges    *MetricFloat64
	Counters  *MetricInt64Sum
	dumpMutex sync.Mutex
	syncWrite bool
	db        *sql.DB
}

func New(args config.ServerConfig, logger *zap.Logger) (*DBStore, error) {
	var ms DBStore

	var err error
	ms.db, err = sql.Open("pgx", args.ConnString)

	if err != nil {

		logger.Info(fmt.Sprintf("Unable to connection to database: %v\n", err))
	}

	ms.syncWrite = args.StoreInterval == 0

	ms.Gauges = NewMetricFloat64()
	ms.Counters = NewMetricInt64Sum()

	ms.Gauges.db = ms.db
	ms.Counters.db = ms.db

	if args.Restore {
		err := ms.Load()
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
	db   *sql.DB
}

func NewMetricFloat64() *MetricFloat64 {
	var v MetricFloat64
	v.data = make(map[string]float64)
	return &v
}

func (ths *MetricFloat64) createTable() error {
	crTableCommand := `CREATE TABLE IF NOT EXISTS public."gauges" 
(
    "Key" text NOT NULL,
    "Value" double precision NOT NULL,
    PRIMARY KEY ("Key")
)`

	_, err := ths.db.Exec(crTableCommand)

	return err
}

func (ths *MetricFloat64) applyValueDB(key string, value float64) error {

	query := "INSERT INTO \"gauges\" (\"Key\", \"Value\") VALUES ('" + key + "'," + fmt.Sprintf("%f", value) + ")"
	query += " ON CONFLICT (\"Key\") DO UPDATE SET \"Value\" = EXCLUDED.\"Value\""
	print(query)
	_, err := ths.db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (ths *MetricFloat64) getValueDB(keys ...string) (map[string]float64, error) {

	var query string
	if len(keys) == 0 {
		query = "SELECT * FROM \"gauges\""
	} else {
		query = "SELECT * FROM \"gauges\" WHERE \"Key\" IN ("
		for i, val := range keys {
			query += ("'" + val + "'")
			if i < len(keys)-1 {
				query += ","
			} else {
				query += ")"
			}
		}
	}

	res := make(map[string]float64)

	rows, err := ths.db.Query(query)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	var (
		key string
		val float64
	)
	for rows.Next() {
		err := rows.Scan(&key, &val)
		if err != nil {
			return nil, err
		}
		res[key] = val
	}
	err = rows.Close()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (ths *MetricFloat64) ReadData(keys ...string) (map[string]float64, error) {

	db, err := ths.getValueDB(keys...)
	if err != nil {
		return nil, err
	}

	return db, nil

}

func (ths *MetricFloat64) WriteData(key string, value string) error {
	v, err := strconv.ParseFloat(value, 64)

	if err == nil {
		ths.mu.Lock()
		defer ths.mu.Unlock()

		err := ths.createTable()
		if err != nil {
			return err
		}

		//ths.data[key] = v
		err = ths.applyValueDB(key, v)
		if err != nil {
			return err
		}

	}

	return nil
}

func (ths *MetricFloat64) WriteDataPP(key string, value float64) error {
	ths.mu.Lock()
	defer ths.mu.Unlock()

	err := ths.createTable()
	if err != nil {
		return err
	}

	err = ths.applyValueDB(key, value)
	if err != nil {
		return err
	}

	//ths.data[key] = value
	return nil
}

// Int64 Cumulative

type MetricInt64Sum struct {
	data map[string]int64
	mu   sync.Mutex
	db   *sql.DB
}

func NewMetricInt64Sum() *MetricInt64Sum {
	var v MetricInt64Sum
	v.data = make(map[string]int64)
	return &v
}

func (ths *MetricInt64Sum) createTable() error {
	crTableCommand := `CREATE TABLE IF NOT EXISTS public."counters"
( 
    "Key" text NOT NULL,
	"Value" bigint NOT NULL,
	PRIMARY KEY ("Key")
)`
	_, err := ths.db.Exec(crTableCommand)

	return err
}

func (ths *MetricInt64Sum) applyValueDB(key string, value int64) error {

	query := "INSERT INTO \"counters\" (\"Key\", \"Value\") VALUES ('" + key + "'," + strconv.Itoa(int(value)) + ")"
	_, err := ths.db.Exec(query)
	if err != nil {
		// Looks like this counter already exist, try UPDATE
		query = "UPDATE \"counters\" SET \"Value\" = \"Value\" + " + strconv.Itoa(int(value)) + " WHERE \"Key\" = '" + key + "'"
		_, err = ths.db.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ths *MetricInt64Sum) getValueDB(keys ...string) (map[string]int64, error) {

	var query string
	if len(keys) == 0 {
		query = "SELECT * FROM \"counters\""
	} else {
		query = "SELECT * FROM \"counters\" WHERE \"Key\" IN ("
		for i, val := range keys {
			query += ("'" + val + "'")
			if i < len(keys)-1 {
				query += ","
			} else {
				query += ")"
			}
		}
	}

	res := make(map[string]int64)

	rows, err := ths.db.Query(query)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	var (
		key string
		val int64
	)
	for rows.Next() {
		err := rows.Scan(&key, &val)
		if err != nil {
			return nil, err
		}
		res[key] = val
	}
	err = rows.Close()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (ths *MetricInt64Sum) ReadData(keys ...string) (map[string]int64, error) {
	ths.mu.Lock()
	defer ths.mu.Unlock()

	db, err := ths.getValueDB(keys...)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (ths *MetricInt64Sum) WriteData(key string, value string) error {

	ths.mu.Lock()
	defer ths.mu.Unlock()

	val, err := strconv.Atoi(value)
	if err != nil {
		return err
	}

	err = ths.createTable()
	if err != nil {
		return err
	}

	//ths.data[key] += v
	err = ths.applyValueDB(key, int64(val))
	if err != nil {
		return err
	}

	return nil
}

func (ths *MetricInt64Sum) WriteDataPP(key string, value int64) error {

	ths.mu.Lock()
	defer ths.mu.Unlock()

	err := ths.createTable()
	if err != nil {
		return err
	}

	//ths.data[key] += v
	err = ths.applyValueDB(key, value)
	if err != nil {
		return err
	}

	return nil
}

// DumpLoad

func (ms *DBStore) Dump() error {

	/*
		mdb := storagecommons.MetricsDB{MetricsDB: make([]storagecommons.Metrics, 0)}

		data, err := ms.Gauges.ReadData()
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

		data2, err := ms.Counters.ReadData()
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
	*/
	return nil
}

func (ms *DBStore) Load() error {

	/*data, err := os.ReadFile(ms.fileName)
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
			ms.Counters.WriteDataPP(v.ID, *v.Delta)
		case "gauge":
			ms.Gauges.WriteDataPP(v.ID, *v.Value)
		}
	}*/

	return nil
}

func (ms *DBStore) WriteData(metrics storagecommons.Metrics) (rMetrics storagecommons.Metrics, rError error) {
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

		data, err := ms.Counters.ReadData()
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
		ms.Dump()
	}

	return
}

func (ms *DBStore) ReadData(metrics storagecommons.Metrics) (storagecommons.Metrics, error) {
	switch metrics.MType {
	case "gauge":
		data, err := ms.Gauges.ReadData()

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
		data, err := ms.Counters.ReadData()

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

func (ms *DBStore) Close() error {
	if ms.db != nil {
		return ms.db.Close()
	}
	return nil
}

func (ms *DBStore) GetGauges() storagecommons.StoragerFloat64 {
	return ms.Gauges
}

func (ms *DBStore) GetCounters() storagecommons.StoragerInt64Sum {
	return ms.Counters
}

func (ms *DBStore) Ping() error {
	if ms.db == nil {
		return errors.New("database connection was not established")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	return ms.db.PingContext(ctx)
}

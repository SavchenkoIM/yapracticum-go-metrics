package dbstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/storage/storagecommons"
)

// Postgres DB storage description (see Storager)
type DBStore struct {
	Gauges         *MetricFloat64
	Counters       *MetricInt64Sum
	syncWrite      bool
	cachedCounters map[string]int64
	cachedGauges   map[string]float64
	db             *sql.DB
}

func New(ctx context.Context, args config.ServerConfig, logger *zap.Logger) (*DBStore, error) {
	var ms DBStore

	logger.Sugar().Infof("Creating database storage...")

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
		err := ms.Load(ctx)
		if err != nil {
			logger.Sugar().Infof("Unable to load data from file: %s", err.Error())
		}
	}

	return &ms, nil
}

// Float64

type MetricFloat64 struct {
	db *sql.DB
}

func NewMetricFloat64() *MetricFloat64 {
	return &MetricFloat64{}
}

func (ths *MetricFloat64) createTable(ctx context.Context, tx *sql.Tx) error {
	crTableCommand := `CREATE TABLE IF NOT EXISTS public."gauges" 
(
    "Key" text NOT NULL,
    "Value" double precision NOT NULL,
    PRIMARY KEY ("Key")
)`

	db := NewTxManager(ths.db, tx)
	_, err := db.ExecContext(ctx, crTableCommand)

	return err
}

func (ths *MetricFloat64) applyValueDB(ctx context.Context, tx *sql.Tx, key string, value float64) error {

	// Fails when executed in transaction if duplicate keys exists
	query := `INSERT INTO "gauges" ("Key", "Value") VALUES ($1, $2) ON CONFLICT ("Key") DO UPDATE SET "Value" = EXCLUDED."Value"`

	db := NewTxManager(ths.db, tx)
	_, err := db.ExecContext(ctx, query, key, value)

	if err != nil {
		return err
	}
	return nil
}

func (ths *MetricFloat64) applyValueDBBatch(ctx context.Context, tx *sql.Tx, data map[string]float64) error {

	if data == nil {
		return nil
	}

	ctr := 0
	paramsStr := make([]string, len(data))
	paramsVals := make([]any, len(data)*2)

	for key, val := range data {
		key, val := key, val
		paramsStr[ctr] = fmt.Sprintf("($%d,$%d)", ctr*2+1, ctr*2+2)
		paramsVals[ctr*2] = key
		paramsVals[ctr*2+1] = val
		ctr++
	}

	if ctr == 0 {
		return nil
	}

	err := ths.createTable(ctx, tx)
	if err != nil {
		return err
	}

	query := `INSERT INTO "gauges" ("Key", "Value") VALUES ` + strings.Join(paramsStr, ",") + " "
	query += `ON CONFLICT ("Key") DO UPDATE SET "Value" = EXCLUDED."Value"`

	db := NewTxManager(ths.db, tx)
	_, err = db.ExecContext(ctx, query, paramsVals...)
	if err != nil {
		println(err.Error())
		return err
	}

	return nil

}

func (ths *MetricFloat64) getValueDB(ctx context.Context, keys ...string) (map[string]float64, error) {

	ks := make([]any, 0)
	for _, v := range keys {
		ks = append(ks, v)
	}

	syms := map[bool]rune{true: ',', false: ')'}
	var query string
	if len(keys) == 0 {
		query = `SELECT * FROM "gauges"`
	} else {
		query = `SELECT * FROM "gauges" WHERE "Key" IN (`
		for i := range keys {
			query += fmt.Sprintf("$%d%c", i+1, syms[i < len(keys)-1])
		}
	}

	res := make(map[string]float64)

	var err error
	rows, err := ths.db.QueryContext(ctx, query, ks...)

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
		err = rows.Scan(&key, &val)
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

func (ths *MetricFloat64) ReadData(ctx context.Context, keys ...string) (map[string]float64, error) {

	db, err := ths.getValueDB(ctx, keys...)
	if err != nil {
		return nil, err
	}

	return db, nil

}

func (ths *MetricFloat64) WriteData(ctx context.Context, key string, value string) error {
	v, err := strconv.ParseFloat(value, 64)

	if err == nil {

		err := ths.createTable(ctx, nil)
		if err != nil {
			return err
		}

		err = ths.applyValueDB(ctx, nil, key, v)
		if err != nil {
			return err
		}

	}

	return nil
}

func (ths *MetricFloat64) writeDataPPTX(ctx context.Context, tx *sql.Tx, key string, value float64) error {

	err := ths.createTable(ctx, nil)
	if err != nil {
		return err
	}

	err = ths.applyValueDB(ctx, tx, key, value)
	if err != nil {
		return err
	}
	return nil
}

func (ths *MetricFloat64) WriteDataPP(ctx context.Context, key string, value float64) error {

	return ths.writeDataPPTX(ctx, nil, key, value)
}

// Int64 Cumulative

type MetricInt64Sum struct {
	db *sql.DB
}

func NewMetricInt64Sum() *MetricInt64Sum {
	return &MetricInt64Sum{}
}

func (ths *MetricInt64Sum) createTable(ctx context.Context, tx *sql.Tx) error {
	crTableCommand := `CREATE TABLE IF NOT EXISTS public."counters"
( 
    "Key" text NOT NULL,
	"Value" bigint NOT NULL,
	PRIMARY KEY ("Key")
)`

	db := NewTxManager(ths.db, tx)
	_, err := db.ExecContext(ctx, crTableCommand)

	return err
}

func (ths *MetricInt64Sum) applyValueDB(ctx context.Context, tx *sql.Tx, key string, value int64) error {

	query := `INSERT INTO "counters" ("Key", "Value") VALUES ($1, $2) ON CONFLICT ("Key") DO UPDATE SET "Value" = "counters"."Value" + EXCLUDED."Value"`

	db := NewTxManager(ths.db, tx)
	_, err := db.ExecContext(ctx, query, key, value)

	if err != nil {
		return err
	}

	return nil
}

func (ths *MetricInt64Sum) applyValueDBBatch(ctx context.Context, tx *sql.Tx, data map[string]int64) error {

	if data == nil {
		return nil
	}

	ctr := 0
	paramsStr := make([]string, len(data))
	paramsVals := make([]any, len(data)*2)

	for key, val := range data {
		key, val := key, val
		paramsStr[ctr] = fmt.Sprintf("($%d,$%d)", ctr*2+1, ctr*2+2)
		paramsVals[ctr*2] = key
		paramsVals[ctr*2+1] = val
		ctr++
	}

	if ctr == 0 {
		return nil
	}

	err := ths.createTable(ctx, tx)
	if err != nil {
		return err
	}

	query := `INSERT INTO "counters" ("Key", "Value") VALUES ` + strings.Join(paramsStr, ",") + " "
	query += `ON CONFLICT ("Key") DO UPDATE SET "Value" = "counters"."Value" + EXCLUDED."Value"`

	db := NewTxManager(ths.db, tx)
	_, err = db.ExecContext(ctx, query, paramsVals...)
	if err != nil {
		return err
	}

	return nil

}

func (ths *MetricInt64Sum) getValueDB(ctx context.Context, keys ...string) (map[string]int64, error) {

	ks := make([]any, 0)
	for _, v := range keys {
		ks = append(ks, v)
	}

	syms := map[bool]rune{true: ',', false: ')'}
	var query string
	if len(keys) == 0 {
		query = `SELECT * FROM "counters"`
	} else {
		query = `SELECT * FROM "counters" WHERE "Key" IN (`
		for i := range keys {
			query += fmt.Sprintf("$%d%c", i+1, syms[i < len(keys)-1])
		}
	}

	res := make(map[string]int64)

	rows, err := ths.db.QueryContext(ctx, query, ks...)

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
		err = rows.Scan(&key, &val)
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

func (ths *MetricInt64Sum) ReadData(ctx context.Context, keys ...string) (map[string]int64, error) {

	db, err := ths.getValueDB(ctx, keys...)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (ths *MetricInt64Sum) WriteData(ctx context.Context, key string, value string) error {

	val, err := strconv.Atoi(value)
	if err != nil {
		return err
	}

	err = ths.createTable(ctx, nil)
	if err != nil {
		return err
	}

	//ths.data[key] += v
	err = ths.applyValueDB(ctx, nil, key, int64(val))
	if err != nil {
		return err
	}

	return nil
}

func (ths *MetricInt64Sum) writeDataPPTX(ctx context.Context, tx *sql.Tx, key string, value int64) error {

	err := ths.createTable(ctx, nil)
	if err != nil {
		return err
	}

	//ths.data[key] += v
	err = ths.applyValueDB(ctx, tx, key, value)
	if err != nil {
		return err
	}

	return nil
}

func (ths *MetricInt64Sum) WriteDataPP(ctx context.Context, key string, value int64) error {
	return ths.writeDataPPTX(ctx, nil, key, value)
}

// DumpLoad

func (ms *DBStore) Dump(ctx context.Context) error {

	return nil
}

func (ms *DBStore) Load(ctx context.Context) error {

	return nil
}

func (ms *DBStore) WriteDataMulti(ctx context.Context, metrics storagecommons.MetricsDB) error {

	return ms.WriteDataMultiBatch(ctx, metrics)

	/*tx, err := ms.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}

	for _, v := range metrics.MetricsDB {
		_, err = ms.writeDataTX(ctx, tx, v)

		if err != nil {
			err = tx.Rollback()
			if err != nil {
				return err
			}
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil*/
}

func (ms *DBStore) writeDataTX(ctx context.Context, tx *sql.Tx, metrics storagecommons.Metrics) (rMetrics storagecommons.Metrics, rError error) {
	rError = nil
	rMetrics = metrics

	switch metrics.MType {
	case "gauge":
		if metrics.Value == nil {
			return metrics, errors.New("no Value data provided")
		}
		ms.Gauges.writeDataPPTX(ctx, tx, metrics.ID, *metrics.Value)
		rMetrics = metrics
	case "counter":
		if metrics.Delta == nil {
			return metrics, errors.New("no Delta data provided")
		}
		ms.Counters.writeDataPPTX(ctx, tx, metrics.ID, *metrics.Delta)

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

func (ms *DBStore) WriteData(ctx context.Context, metrics storagecommons.Metrics) (rMetrics storagecommons.Metrics, rError error) {
	return ms.writeDataTX(ctx, nil, metrics)
}

func (ms *DBStore) ReadData(ctx context.Context, metrics storagecommons.Metrics) (storagecommons.Metrics, error) {
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
			return metrics, err
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

func (ms *DBStore) Close(ctx context.Context) error {
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

func (ms *DBStore) Ping(ctx context.Context) error {
	if ms.db == nil {
		return errors.New("database connection was not established")
	}
	ctxw, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
	defer cancel()
	return ms.db.PingContext(ctxw)
}

// Batch write Raw (does not pass github tests)
func (ms *DBStore) WriteDataMultiBatchRaw(ctx context.Context, gauges map[string]float64, counters map[string]int64) error {

	tx, _ := ms.db.BeginTx(ctx, nil)

	ms.Gauges.applyValueDBBatch(ctx, tx, gauges)
	ms.Counters.applyValueDBBatch(ctx, tx, counters)

	err := tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// Batch write (does not pass github tests)
func (ms *DBStore) WriteDataMultiBatch(ctx context.Context, metrics storagecommons.MetricsDB) error {

	gauges := map[string]float64{}
	counters := map[string]int64{}

	for _, val := range metrics.MetricsDB {
		val := val
		switch val.MType {
		case "counter":
			counters[val.ID] += *val.Delta
		case "gauge":
			gauges[val.ID] = *val.Value
		default:
			return errors.New("Unknown metric type: " + val.MType)
		}
	}

	err := ms.WriteDataMultiBatchRaw(ctx, gauges, counters)
	if err != nil {
		return err
	}

	return nil
}

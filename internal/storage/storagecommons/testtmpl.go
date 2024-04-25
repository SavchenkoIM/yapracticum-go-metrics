package storagecommons

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func PerformStoragerTest(t *testing.T, db Storager) {
	ctx := context.Background()
	var m Metrics

	t.Run("Init Counter", func(t *testing.T) {
		m.ID = "testCounter"
		m.MType = "counter"
		var d int64 = 1
		m.Delta = &d
		_, err := db.WriteData(ctx, m)
		assert.NoError(t, err)
	})

	t.Run("Check Counter value", func(t *testing.T) {
		ctr, err := db.GetCounters().ReadData(ctx, "testCounter")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), ctr["testCounter"])
	})

	t.Run("Init Gauge", func(t *testing.T) {
		m.ID = "testGauge"
		m.MType = "gauge"
		var v = 5.5
		m.Value = &v
		_, err := db.WriteData(ctx, m)
		assert.NoError(t, err)
	})

	t.Run("Check Gauge value", func(t *testing.T) {
		ctr, err := db.GetGauges().ReadData(ctx, "testGauge")
		assert.NoError(t, err)
		assert.Equal(t, 5.5, ctr["testGauge"])
	})

	var mdb MetricsDB
	var f float64 = 6.6
	var i int64 = 1
	mdb.MetricsDB = []Metrics{
		{ID: "gm1", MType: "gauge", Value: &f},
		{ID: "cm1", MType: "counter", Delta: &i},
	}
	t.Run("Multiple Write", func(t *testing.T) {
		err := db.WriteDataMulty(ctx, mdb)
		assert.NoError(t, err)
	})

	t.Run("Check Gauge Value", func(t *testing.T) {
		m := Metrics{ID: "gm1", MType: "gauge"}
		data, err := db.ReadData(ctx, m)
		assert.NoError(t, err)
		assert.Equal(t, 6.6, *data.Value)
	})

	t.Run("Check Counter Value", func(t *testing.T) {
		m := Metrics{ID: "cm1", MType: "counter"}
		data, err := db.ReadData(ctx, m)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), *data.Delta)
	})

	t.Run("Read Value Of Unknown Type", func(t *testing.T) {
		m := Metrics{ID: "cm1", MType: "countter"}
		_, err := db.ReadData(ctx, m)
		assert.Error(t, err)
	})

	t.Run("Write Value Of Unknown Type", func(t *testing.T) {
		m := Metrics{ID: "cm1", MType: "countter"}
		_, err := db.WriteData(ctx, m)
		assert.Error(t, err)
	})

	t.Run("Read Not Existing Counter", func(t *testing.T) {
		m := Metrics{ID: "cm18", MType: "counter"}
		_, err := db.ReadData(ctx, m)
		assert.Error(t, err)
	})

	t.Run("Read Not Existing Gauge", func(t *testing.T) {
		m := Metrics{ID: "gm18", MType: "gauge"}
		_, err := db.ReadData(ctx, m)
		assert.Error(t, err)
	})
}

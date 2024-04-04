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

	t.Run("GetCounters Pass Two Keys", func(t *testing.T) {
		_, err := db.GetCounters().ReadData(ctx, "dwd", "fewfw")
		assert.Error(t, err)
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

	t.Run("GetGauges Pass Two Keys", func(t *testing.T) {
		_, err := db.GetGauges().ReadData(ctx, "dwd", "fewfw")
		assert.Error(t, err)
	})

	t.Run("Dump Data To Disc", func(t *testing.T) {
		err := db.Dump(ctx)
		assert.NoError(t, err)
	})

	t.Run("Increment Counter", func(t *testing.T) {
		m.ID = "testCounter"
		m.MType = "counter"
		var d int64 = 1
		m.Delta = &d
		_, err := db.WriteData(ctx, m)
		assert.NoError(t, err)
	})

	t.Run("ReCheck Counter Value Aft Inc", func(t *testing.T) {
		ctr, err := db.GetCounters().ReadData(ctx, "testCounter")
		assert.NoError(t, err)
		assert.Equal(t, int64(2), ctr["testCounter"])
	})

	t.Run("Load Data From Disc", func(t *testing.T) {
		err := db.Load(ctx)
		assert.NoError(t, err)
	})

	// Bug/Feature: Loading of Counter increments existing counter instead of replacing it
	// Load is to be called only when database is empty
	t.Run("ReCheck Counter Value Aft Load", func(t *testing.T) {
		ctr, err := db.GetCounters().ReadData(ctx, "testCounter")
		assert.NoError(t, err)
		assert.Equal(t, int64(3), ctr["testCounter"])
	})
}

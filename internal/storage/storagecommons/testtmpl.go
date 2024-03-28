package storagecommons

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func PerformStoregerTest(t *testing.T, db Storager) {
	ctx := context.Background()
	var m Metrics

	t.Run("Init Counter", func(t *testing.T) {
		m.ID = "testCounter"
		m.MType = "counter"
		var d int64 = 1
		m.Delta = &d
		db.WriteData(ctx, m)

		ctr, err := db.GetCounters().ReadData(ctx, "testCounter")
		assert.NoError(t, err)
		assert.Equal(t, ctr["testCounter"], d)
	})

	t.Run("Init Gauge", func(t *testing.T) {
		m.ID = "testGauge"
		m.MType = "gauge"
		var v float64 = 5.5
		m.Value = &v
		db.WriteData(ctx, m)

		ctr, err := db.GetGauges().ReadData(ctx, "testGauge")
		assert.NoError(t, err)
		assert.Equal(t, ctr["testGauge"], v)
	})
}

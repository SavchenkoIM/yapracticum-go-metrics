package filestore

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"os"
	"testing"
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/storage/storagecommons"
	"yaprakticum-go-track2/internal/testhelpers"
)

func Test(t *testing.T) {
	ctx := context.Background()
	logger := testhelpers.GetCustomZap(zap.ErrorLevel)
	db, err := New(ctx, config.ServerConfig{StoreInterval: 3000, FileStoragePath: "test.json"}, logger)
	assert.NoError(t, err)
	storagecommons.PerformStoragerTest(t, db)

	var m storagecommons.Metrics

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

	t.Run("ReCheck Counter Value Aft Load", func(t *testing.T) {
		ctr, err := db.GetCounters().ReadData(ctx, "testCounter")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), ctr["testCounter"])
	})

	os.Remove("test.json")
}

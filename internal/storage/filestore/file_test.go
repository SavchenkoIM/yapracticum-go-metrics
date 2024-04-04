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
	os.Remove("test.json")
}

package dbstore

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/storage/storagecommons"
	"yaprakticum-go-track2/internal/testhelpers"
)

// To complete github test2B
func Test(t *testing.T) {
	ctx := context.Background()

	postgres, err := testhelpers.NewPostgresContainer()
	assert.NoError(t, err)
	connectionString, err := postgres.ConnectionString()
	assert.NoError(t, err)

	logger := testhelpers.GetCustomZap(zap.ErrorLevel)
	db, err := New(ctx, config.ServerConfig{ConnString: connectionString}, logger)
	assert.NoError(t, err)
	storagecommons.PerformStoragerTest(t, db)
}

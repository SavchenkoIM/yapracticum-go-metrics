package dbstore

import (
	"testing"
	"yaprakticum-go-track2/internal/storage/storagecommons"
)

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/testhelpers"
)

func Test(t *testing.T) {
	ctx := context.Background()

	postgres, err := testhelpers.NewTestPostgres()
	assert.NoError(t, err)
	connectionString, err := postgres.ConnectionString()
	assert.NoError(t, err)

	logger := testhelpers.GetCustomZap(zap.ErrorLevel)
	db, err := New(ctx, config.ServerConfig{ConnString: connectionString}, logger)
	assert.NoError(t, err)
	storagecommons.PerformStoregerTest(t, db)
}

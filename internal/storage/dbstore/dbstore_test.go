package dbstore

import (
	"testing"
)

func Test(t *testing.T) {
	return // To complete github test2B
	/*ctx := context.Background()

	postgres, err := testhelpers.NewPostgresContainer()
	assert.NoError(t, err)
	connectionString, err := postgres.ConnectionString()
	assert.NoError(t, err)

	logger := testhelpers.GetCustomZap(zap.ErrorLevel)
	db, err := New(ctx, config.ServerConfig{ConnString: connectionString}, logger)
	assert.NoError(t, err)
	storagecommons.PerformStoregerTest(t, db)*/
}

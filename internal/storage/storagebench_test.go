package storage

import (
	"context"
	"go.uber.org/zap"
	"testing"
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/storage/storagecommons"
	"yaprakticum-go-track2/internal/testhelpers"
)

func BenchmarkPostgres(b *testing.B) {
	postgres, err := testhelpers.NewPostgresContainer()
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	logger := testhelpers.GetCustomZap(zap.ErrorLevel)
	connectionString, err := postgres.ConnectionString()
	if err != nil {
		b.Fatal(err)
	}
	db, _ := InitStorage(ctx, config.ServerConfig{
		ConnString: connectionString,
	}, logger)
	b.ResetTimer()

	defer func(postgres *testhelpers.PostgresContainer) {
		b.StopTimer()
		postgres.Close()
		b.StartTimer()
	}(postgres)

	var val float64 = 20
	var m storagecommons.Metrics
	m.ID = "g1"
	m.MType = "gauge"
	m.Value = &val

	for i := 0; i < b.N; i++ {
		db.WriteData(ctx, m)
	}
}

func BenchmarkInMemory(b *testing.B) {
	ctx := context.Background()
	logger := testhelpers.GetCustomZap(zap.ErrorLevel)
	db, _ := InitStorage(ctx, config.ServerConfig{}, logger)
	b.ResetTimer()

	var val float64 = 20
	var m storagecommons.Metrics
	m.ID = "g1"
	m.MType = "gauge"
	m.Value = &val

	for i := 0; i < b.N; i++ {
		db.WriteData(ctx, m)
	}
}

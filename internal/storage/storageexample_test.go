package storage

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/storage/storagecommons"
	"yaprakticum-go-track2/internal/testhelpers"
)

func ExampleInitStorage() {
	ctx := context.Background()
	storage, _ := InitStorage(ctx, config.ServerConfig{}, testhelpers.GetCustomZap(zap.ErrorLevel))
	f := float64(5)
	storage.WriteData(ctx, storagecommons.Metrics{ID: "testGauge", Value: &f, MType: "gauge"})
	data, _ := storage.GetGauges().ReadData(ctx)
	fmt.Printf("%v", data)
	// Output:
	//map[testGauge:5]
}

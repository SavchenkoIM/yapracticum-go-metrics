package handlers

import (
	"bytes"
	"context"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/storage"
	"yaprakticum-go-track2/internal/testhelpers"
)

func performBench(b *testing.B, h Handlers) {
	reqData := `{ "id": "g1", "value": 5, "type": "gauge" }`
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		res := httptest.NewRecorder()
		req.Body = io.NopCloser(bytes.NewBuffer([]byte(reqData)))
		req.ContentLength = int64(len(reqData))
		h.MetricsUpdateHandlerREST(res, req)
	}
}

func BenchmarkAddGaugeInMemory(b *testing.B) {
	ctx := context.Background()
	logger := testhelpers.GetCustomZap(zap.ErrorLevel)
	dataStorage, _ := storage.InitStorage(ctx, config.ServerConfig{}, logger)
	h := NewHandlers(dataStorage, config.ServerConfig{})
	b.ResetTimer()

	performBench(b, h)
}

func BenchmarkAddGaugePostgres(b *testing.B) {
	postgres, err := testhelpers.NewPostgresContainer()
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	logger := testhelpers.GetCustomZap(zap.ErrorLevel)
	connectionString, _ := postgres.ConnectionString()
	dataStorage, _ := storage.InitStorage(ctx, config.ServerConfig{ConnString: connectionString}, logger)
	h := NewHandlers(dataStorage, config.ServerConfig{})
	b.ResetTimer()

	defer func(postgres *testhelpers.PostgresContainer) {
		b.StopTimer()
		postgres.Close()
		b.StartTimer()
	}(postgres)

	performBench(b, h)
}

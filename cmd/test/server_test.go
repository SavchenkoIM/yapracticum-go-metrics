package test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/handlers"
	"yaprakticum-go-track2/internal/prom"
	"yaprakticum-go-track2/internal/shared"
	"yaprakticum-go-track2/internal/storage"
)

var once sync.Once
var cpm *prom.CustomPromMetrics

func performTest(t *testing.T, db *storage.Storage) {

	type kv struct {
		typ   string
		key   string
		value interface{}
	}

	tests := []struct {
		testName       string
		method         string
		url            string
		wantStatusCode int
		wantKv         []kv
	}{
		{testName: "GET Request", method: http.MethodGet, url: "/update/counter/testVal/1", wantStatusCode: http.StatusMethodNotAllowed, wantKv: nil},
		//{testName: "Without \"update\" prefix", method: http.MethodPost, url: "/counter/testVal/1", wantStatusCode: http.StatusNotFound, wantKv: nil},
		{testName: "Incorrect type", method: http.MethodPost, url: "/update/countter/testVal/1", wantStatusCode: http.StatusBadRequest, wantKv: nil},
		{testName: "Incorrect value", method: http.MethodPost, url: "/update/counter/testVal/f1", wantStatusCode: http.StatusBadRequest, wantKv: nil},
		{testName: "No value", method: http.MethodPost, url: "/update/counter/testVal", wantStatusCode: http.StatusBadRequest, wantKv: nil},
		{testName: "No name", method: http.MethodPost, url: "/update/counter/", wantStatusCode: http.StatusNotFound, wantKv: nil},
		{testName: "No type", method: http.MethodPost, url: "/update/", wantStatusCode: http.StatusBadRequest, wantKv: nil},
		{testName: "Initializing counter testVal", method: http.MethodPost, url: "/update/counter/testVal/1", wantStatusCode: http.StatusOK, wantKv: []kv{{typ: "counter", key: "testVal", value: int64(1)}}},
		{testName: "Adding value to existing counter testVal", method: http.MethodPost, url: "/update/counter/testVal/2", wantStatusCode: http.StatusOK, wantKv: []kv{{typ: "counter", key: "testVal", value: int64(3)}}},
		{testName: "Initializing gauge testVal", method: http.MethodPost, url: "/update/gauge/testVal/1", wantStatusCode: http.StatusOK, wantKv: []kv{{typ: "gauge", key: "testVal", value: float64(1)}}},
		{testName: "Setting value to existing gauge testVal", method: http.MethodPost, url: "/update/gauge/testVal/2", wantStatusCode: http.StatusOK, wantKv: []kv{{typ: "gauge", key: "testVal", value: float64(2)}}},
	}

	ctx := context.Background()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	shared.Logger = logger

	once.Do(func() {
		cpm = prom.NewCustomPromMetrics()
	})
	srv := httptest.NewServer(handlers.Router(handlers.NewHandlers(db, config.ServerConfig{}), cpm))
	defer srv.Close()

	for _, tt := range tests {

		t.Run(tt.testName, func(t *testing.T) {

			var res *http.Response

			if tt.method == http.MethodGet {
				res, _ = srv.Client().Get(srv.URL + tt.url)
				res.Body.Close()
			} else {
				res, _ = srv.Client().Post(srv.URL+tt.url, "text/plain", nil)
				res.Body.Close()
			}

			assert.Equal(t, tt.wantStatusCode, res.StatusCode)
			for _, v := range tt.wantKv {

				switch v.typ {
				case "gauge":
					val, _ := db.GetGauges().ReadData(ctx, v.key)
					assert.Equal(t, v.value, val[v.key])
				case "counter":
					val, _ := db.GetCounters().ReadData(ctx, v.key)
					assert.Equal(t, v.value, val[v.key])
				}

			}
		})
	}
}

func TestInMemory(t *testing.T) {
	var ctx context.Context
	z, _ := zap.NewDevelopment()
	db, _ := storage.InitStorage(ctx, config.ServerConfig{}, z)
	performTest(t, db)
}

// To complete github test2B
/*func TestPostgres(t *testing.T) {
	postgres, err := testhelpers.NewPostgresContainer()
	if err != nil {
		t.Fatal(err)
	}
	defer postgres.Close()

	ctx := context.Background()
	connectionString, err := postgres.ConnectionString()
	if err != nil {
		t.Fatal(err)
	}
	z, _ := zap.NewDevelopment()
	db, _ := storage.InitStorage(ctx, config.ServerConfig{ConnString: connectionString}, z)
	performTest(t, db)
}*/

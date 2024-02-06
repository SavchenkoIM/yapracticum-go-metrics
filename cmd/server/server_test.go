package main

import (
	"context"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/handlers/getmetrics"
	"yaprakticum-go-track2/internal/handlers/updatemetrics"
	"yaprakticum-go-track2/internal/shared"
	"yaprakticum-go-track2/internal/storage"

	"github.com/stretchr/testify/assert"
)

func TestIter2Server(t *testing.T) {

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

	var ctx context.Context

	z, _ := zap.NewDevelopment()
	db, _ := storage.InitStorage(ctx, config.ServerConfig{}, z)
	updatemetrics.SetDataStorage(db)
	getmetric.SetDataStorage(db)
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	shared.Logger = logger

	srv := httptest.NewServer(Router())
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

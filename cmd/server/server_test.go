package main

import (
	_ "fmt"
	_ "io"
	"net/http"
	"net/http/httptest"
	"testing"
	"yaprakticum-go-track2/internal/handlers/getMetrics"
	"yaprakticum-go-track2/internal/handlers/updateMetrics"
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
		testName        string
		method          string
		url             string
		want_statusCode int
		want_kv         []kv
	}{
		{testName: "GET Request", method: http.MethodGet, url: "/update/counter/testVal/1", want_statusCode: http.StatusBadRequest, want_kv: nil},
		//{testName: "Without \"update\" prefix", method: http.MethodPost, url: "/counter/testVal/1", want_statusCode: http.StatusNotFound, want_kv: nil},
		{testName: "Incorrect type", method: http.MethodPost, url: "/update/countter/testVal/1", want_statusCode: http.StatusBadRequest, want_kv: nil},
		{testName: "Incorrect value", method: http.MethodPost, url: "/update/counter/testVal/f1", want_statusCode: http.StatusBadRequest, want_kv: nil},
		{testName: "No value", method: http.MethodPost, url: "/update/counter/testVal/", want_statusCode: http.StatusBadRequest, want_kv: nil},
		{testName: "No key", method: http.MethodPost, url: "/update/counter/", want_statusCode: http.StatusNotFound, want_kv: nil},
		{testName: "No type", method: http.MethodPost, url: "/update/", want_statusCode: http.StatusNotFound, want_kv: nil},
		{testName: "Initializing counter testVal", method: http.MethodPost, url: "/update/counter/testVal/1", want_statusCode: http.StatusOK, want_kv: []kv{kv{typ: "counter", key: "testVal", value: int64(1)}}},
		{testName: "Adding value to existing counter testVal", method: http.MethodPost, url: "/update/counter/testVal/2", want_statusCode: http.StatusOK, want_kv: []kv{kv{typ: "counter", key: "testVal", value: int64(3)}}},
		{testName: "Initializing gauge testVal", method: http.MethodPost, url: "/update/gauge/testVal/1", want_statusCode: http.StatusOK, want_kv: []kv{kv{typ: "gauge", key: "testVal", value: float64(1)}}},
		{testName: "Setting value to existing gauge testVal", method: http.MethodPost, url: "/update/gauge/testVal/2", want_statusCode: http.StatusOK, want_kv: []kv{kv{typ: "gauge", key: "testVal", value: float64(2)}}},
	}

	db := storage.InitStorage()
	updateMetrics.SetDataStorage(&db)
	getMetric.SetDataStorage(&db)

	srv := httptest.NewServer(Router())
	defer srv.Close()

	for _, tt := range tests {

		//if t != nil {
		t.Run(tt.testName, func(t *testing.T) {

			var res *http.Response

			if tt.method == http.MethodGet {
				res, _ = srv.Client().Get(srv.URL + tt.url)
			} else {
				res, _ = srv.Client().Post(srv.URL+tt.url, "text/plain", nil)
			}

			assert.Equal(t, tt.want_statusCode, res.StatusCode)
			if tt.want_kv != nil {
				for _, v := range tt.want_kv {
					val, _ := db.ReadData(v.typ, v.key)
					assert.Equal(t, v.value, val)
				}
			}
		})
		/*} else {

			//var res *http.Response

			if tt.method == http.MethodGet {
				res, _ := srv.Client().Get(srv.URL + tt.url)
				defer res.Body.Close()
				print(res.StatusCode)
			} else {
				res, _ := srv.Client().Post(srv.URL+tt.url, "text/plain", nil)
				res.Body.Close()
				print(res.StatusCode)
			}

			//println(&res.StatusCode)

		}*/
	}

	/*	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, nil)
			//w := httptest.NewRecorder()
			//http.StripPrefix("/update/", http.HandlerFunc(updateMetrics.MetricUpdateHandler)).ServeHTTP(w, req)
			//http.HandlerFunc(updateMetrics.MetricUpdateHandler).ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()
			body, _ := io.ReadAll(res.Body)
			fmt.Printf("%s\n", string(body))
			assert.Equal(t, tt.want_statusCode, res.StatusCode)
			if tt.want_kv != nil {
				for _, v := range tt.want_kv {
					val, _ := db.ReadData(v.typ, v.key)
					assert.Equal(t, v.value, val)
				}
			}
		})
	}*/

}

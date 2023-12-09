package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"yaprakticum-go-track2/internal/handlers/getmetrics"
	"yaprakticum-go-track2/internal/handlers/updatemetrics"
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
		{testName: "GET Request", method: http.MethodGet, url: "/update/counter/testVal/1", wantStatusCode: http.StatusBadRequest, wantKv: nil},
		//{testName: "Without \"update\" prefix", method: http.MethodPost, url: "/counter/testVal/1", wantStatusCode: http.StatusNotFound, wantKv: nil},
		{testName: "Incorrect type", method: http.MethodPost, url: "/update/countter/testVal/1", wantStatusCode: http.StatusBadRequest, wantKv: nil},
		{testName: "Incorrect value", method: http.MethodPost, url: "/update/counter/testVal/f1", wantStatusCode: http.StatusBadRequest, wantKv: nil},
		{testName: "No value", method: http.MethodPost, url: "/update/counter/testVal/", wantStatusCode: http.StatusBadRequest, wantKv: nil},
		{testName: "No key", method: http.MethodPost, url: "/update/counter/", wantStatusCode: http.StatusNotFound, wantKv: nil},
		{testName: "No type", method: http.MethodPost, url: "/update/", wantStatusCode: http.StatusNotFound, wantKv: nil},
		{testName: "Initializing counter testVal", method: http.MethodPost, url: "/update/counter/testVal/1", wantStatusCode: http.StatusOK, wantKv: []kv{kv{typ: "counter", key: "testVal", value: int64(1)}}},
		{testName: "Adding value to existing counter testVal", method: http.MethodPost, url: "/update/counter/testVal/2", wantStatusCode: http.StatusOK, wantKv: []kv{kv{typ: "counter", key: "testVal", value: int64(3)}}},
		{testName: "Initializing gauge testVal", method: http.MethodPost, url: "/update/gauge/testVal/1", wantStatusCode: http.StatusOK, wantKv: []kv{kv{typ: "gauge", key: "testVal", value: float64(1)}}},
		{testName: "Setting value to existing gauge testVal", method: http.MethodPost, url: "/update/gauge/testVal/2", wantStatusCode: http.StatusOK, wantKv: []kv{kv{typ: "gauge", key: "testVal", value: float64(2)}}},
	}

	db := storage.InitStorage()
	updatemetrics.SetDataStorage(&db)
	getmetric.SetDataStorage(&db)

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
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatusCode, res.StatusCode)
			if tt.wantKv != nil {
				for _, v := range tt.wantKv {
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
			assert.Equal(t, tt.wantStatusCode, res.StatusCode)
			if tt.wantKv != nil {
				for _, v := range tt.wantKv {
					val, _ := db.ReadData(v.typ, v.key)
					assert.Equal(t, v.value, val)
				}
			}
		})
	}*/

}

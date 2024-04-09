package middleware

import (
	"net/http"
	"time"
	"yaprakticum-go-track2/internal/prom"
)

// Custom prom metrics handling
func Prom(m *prom.CustomPromMetrics) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m.IncHttpRequest()
			ts := time.Now()
			var erw = extResponseWriter{WrittenDataLength: 0, StatusCode: 200, ResponseWriter: w}
			h.ServeHTTP(&erw, r)
			m.IncHttpHistogram(r.Method, r.RequestURI, erw.StatusCode, time.Now().Sub(ts).Seconds())
		})
	}
}

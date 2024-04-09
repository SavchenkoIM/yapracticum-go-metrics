package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"
	"yaprakticum-go-track2/internal/shared"
)

// Middleware function that writes log of every chained request
func WithLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sugar := shared.Logger.Sugar()
		ts := time.Now()

		b := make([]byte, r.ContentLength)
		r.Body.Read(b)
		r.Body.Close()
		r.Body = io.NopCloser(bytes.NewBuffer(b))

		var erw = extResponseWriter{WrittenDataLength: 0, StatusCode: 200, ResponseWriter: w}
		h.ServeHTTP(&erw, r)

		sugar.Infof("URI: %s, Method: %s, Body: %s, Runtime: %d msec, RespStatusCode: %d, RespDataLen: %d",
			r.RequestURI, r.Method, string(b), time.Since(ts).Milliseconds(),
			erw.StatusCode, erw.WrittenDataLength)

	})
}

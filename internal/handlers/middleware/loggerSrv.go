package middleware

import (
	"go.uber.org/zap"
	"net/http"
	"time"
)

var logger *zap.Logger

func SetLogger(sLogger *zap.Logger) {
	logger = sLogger
}

type extResponseWriter struct {
	http.ResponseWriter
	StatusCode        int
	WrittenDataLength int
}

func (erw *extResponseWriter) Write(data []byte) (int, error) {
	ln, err := erw.ResponseWriter.Write(data)
	if err == nil {
		erw.WrittenDataLength += ln
	}
	return ln, err
}

func (erw *extResponseWriter) Header() http.Header {
	return erw.ResponseWriter.Header()
}

func (erw *extResponseWriter) WriteHeader(statusCode int) {
	erw.StatusCode = statusCode
	erw.ResponseWriter.WriteHeader(statusCode)
}

func WithLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sugar := logger.Sugar()
		ts := time.Now()

		var erw = extResponseWriter{WrittenDataLength: 0, StatusCode: 200, ResponseWriter: w}
		h.ServeHTTP(&erw, r)

		sugar.Infof("URI: %s, Method: %s, Runtime: %d msec, RespStatusCode: %d, RespDataLen: %d",
			r.RequestURI, r.Method, time.Since(ts).Milliseconds(),
			erw.StatusCode, erw.WrittenDataLength)
	})
}

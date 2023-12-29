package middleware

import (
	"bytes"
	"go.uber.org/zap"
	"io"
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

		b := make([]byte, r.ContentLength)
		r.Body.Read(b)
		r.Body.Close()
		r.Body = io.NopCloser(bytes.NewBuffer(b))

		var erw = extResponseWriter{WrittenDataLength: 0, StatusCode: 200, ResponseWriter: w}
		h.ServeHTTP(&erw, r)

		sugar.Infof("URI: %s, Method: %s, Body: %s, Runtime: %d msec, RespStatusCode: %d, RespDataLen: %d",
			r.RequestURI, r.Method, string(b), time.Since(ts).Milliseconds(),
			erw.StatusCode, erw.WrittenDataLength)

		/*f, _ := os.OpenFile("D:\\log.txt", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 644)
		defer f.Close()
		tn := time.Now()
		f.WriteString(strconv.Itoa(tn.Minute()) + ":" + strconv.Itoa(tn.Second()) + "\t" + r.RequestURI + "\t\t" + string(b) + "\n")*/

	})
}

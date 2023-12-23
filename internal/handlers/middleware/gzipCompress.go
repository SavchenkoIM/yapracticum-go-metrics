package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type customResponseWriter struct {
	http.ResponseWriter
	w io.Writer
}

func (crw customResponseWriter) Write(b []byte) (int, error) {
	return crw.w.Write(b)
}

func GzipHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		acceptGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
		encodedGzip := strings.Contains(r.Header.Get("Content-Encoding"), "gzip")

		wh := w

		if encodedGzip {
			gr, _ := gzip.NewReader(r.Body)
			bodyData := bytes.Buffer{}
			_, err := bodyData.ReadFrom(gr)
			if err != nil {
				http.Error(w, "GZIP decompression error", 500)
			}
			r.Body = io.NopCloser(bytes.NewReader(bodyData.Bytes()))
			r.ContentLength = int64(len(bodyData.Bytes()))
		}

		if acceptGzip {
			gz := gzip.NewWriter(w)
			defer gz.Close()
			wh = customResponseWriter{ResponseWriter: w, w: gz}
			w.Header().Set("Content-Encoding", "gzip")
		}

		h.ServeHTTP(wh, r)

	})
}

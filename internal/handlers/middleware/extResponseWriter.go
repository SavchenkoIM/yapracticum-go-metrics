package middleware

import "net/http"

// ResponseWriter with extended functionality: saving status code and written data length
type extResponseWriter struct {
	http.ResponseWriter
	StatusCode        int
	WrittenDataLength int
}

// Overloaded ResponseWriter's Write method
func (erw *extResponseWriter) Write(data []byte) (int, error) {
	ln, err := erw.ResponseWriter.Write(data)
	if err == nil {
		erw.WrittenDataLength += ln
	}
	return ln, err
}

// Overloaded ResponseWriter's Header method
func (erw *extResponseWriter) Header() http.Header {
	return erw.ResponseWriter.Header()
}

// Overloaded ResponseWriter's WriteHeader method
func (erw *extResponseWriter) WriteHeader(statusCode int) {
	erw.StatusCode = statusCode
	erw.ResponseWriter.WriteHeader(statusCode)
}

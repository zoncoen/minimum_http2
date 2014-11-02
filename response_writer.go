package minimum_http2

import (
	"bytes"
	"net/http"
)

type ResponseWriter struct {
	Status  int
	Headers http.Header
	Body    *bytes.Buffer
}

func NewResponseWriter() *ResponseWriter {
	return &ResponseWriter{
		Status:  0,
		Headers: make(http.Header),
		Body:    new(bytes.Buffer),
	}
}

func (rw *ResponseWriter) Header() http.Header {
	return rw.Headers
}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	if rw.Status == 0 {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.Body.Write(b)
}

func (rw *ResponseWriter) WriteHeader(Status int) {
	rw.Status = Status
}

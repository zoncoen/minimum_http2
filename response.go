package minimum_http2

import (
	"net/http"
)

type Response struct {
	Header http.Header
	Body   []byte
}

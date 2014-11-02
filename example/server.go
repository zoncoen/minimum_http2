package main

import (
	"fmt"
	http2 "github.com/zoncoen/minimum_http2"
	"net/http"
	"os"
)

var serveMux = http.NewServeMux()

func main() {
	args := os.Args
	if len(args) != 2 {
		fmt.Println("Usage: [DEBUG=1] go run filename {PORT}")
		os.Exit(1)
	}

	http2.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello HTTP/2!")
	})

	http2.ListenAndServe(args[1])
}

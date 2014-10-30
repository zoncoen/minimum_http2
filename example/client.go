package main

import (
	"fmt"
	http2 "github.com/zoncoen/minimum_http2"
	"os"
	"time"
)

func main() {
	// URL, _ := url.Parse("192.168.59.103:80")
	transport := &http2.Transport{}

	err := transport.Connect("127.0.01:5050")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	stream := transport.NewStream()
	time.Sleep(2 * time.Second)
	frame := http2.NewHeadersFrame(http2.UNSET, 0x1)
	frame.AddHeader(":method", "GET")
	frame.AddHeader(":scheme", "http")
	frame.AddHeader(":authority", transport.Conn.RemoteAddr().String())
	frame.AddHeader(":path", "/")
	stream.WriteChan <- frame
	time.Sleep(2 * time.Second)
}

package minimum_http2

import (
	"net"
	"net/http"
)

var serveMux = http.NewServeMux()

func HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	serveMux.HandleFunc(pattern, handler)
}

func Serve(conn net.Conn) {
	transport := &Transport{}
	transport.Conn = conn

	preface := make([]byte, len(CONNECTION_PREFACE))
	_, err := transport.Conn.Read(preface)
	if err != nil {
		debug(err.Error())
	}
	if string(preface) != CONNECTION_PREFACE {
		debug("Error: recv invalid connection preface")
	}
	if string(preface) == CONNECTION_PREFACE {
		debug("Recv connection preface")

		transport.WriteChan = make(chan Frame, 256)
		transport.NextId = 2
		transport.Streams = map[uint32]*Stream{}
		transport.Streams[0] = NewStream(0, transport.WriteChan)

		go transport.ServerReadLoop(serveMux)
		go transport.WriteLoop()

		transport.WriteEmptySettings()
	}
}

func ListenAndServe(port string) {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		panic(err.Error())
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			debug(err.Error())
			continue
		}

		go Serve(conn)
	}

}

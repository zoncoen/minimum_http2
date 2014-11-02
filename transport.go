package minimum_http2

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

type Transport struct {
	Conn      net.Conn
	WriteChan chan Frame
	NextId    uint32
	Streams   map[uint32]*Stream
	Quit      chan struct{}
}

func (transport *Transport) Connect(addr string) (err error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	transport.Conn = conn
	transport.WriteChan = make(chan Frame, 256)
	transport.NextId = 1
	transport.Streams = map[uint32]*Stream{}
	transport.Streams[0] = NewStream(0, transport.WriteChan)

	quit := make(chan error)

	// read server settings frame and ACK
	go func() {
		var err error = nil
		settings, ack := 0, 0
		for settings+ack < 2 {
			header := new(FrameHeader)
			err = binary.Read(transport.Conn, binary.BigEndian, header)
			if err != nil {
				break
			}

			switch {
			case header.Type == SettingsFrameType:
				debug("Recv", header.TypeString(), header.FlagsString(), header.StreamId)
				if header.Flags == UNSET {
					ackFrame := NewSettingsFrame(ACK)
					ackFrame.Write(transport.Conn)
					header = ackFrame.Header()
					debug("Send", header.TypeString(), header.FlagsString(), header.StreamId)
					settings = 1
				} else if header.Flags == ACK {
					ack = 1
				}
			default:
			}
		}
		quit <- err
	}()

	err = transport.WriteConnectionPreface()
	if err != nil {
		return err
	}
	err = transport.WriteEmptySettings()
	if err != nil {
		return err
	}

	// wait recv server settings frame and recv settings frame ACK
	err = <-quit
	if err != nil {
		return err
	}

	go transport.ReadLoop()
	go transport.WriteLoop()

	return nil
}

func (transport *Transport) Close() {
	for _, v := range transport.Streams {
		v.Quit <- struct{}{}
	}
	transport.Conn.Close()
}

func (transport *Transport) WriteConnectionPreface() (err error) {
	debug("Send connection preface")
	_, err = transport.Conn.Write([]byte(CONNECTION_PREFACE))
	if err != nil {
		return err
	}

	return nil
}

func (transport *Transport) WriteEmptySettings() (err error) {
	frame := NewSettingsFrame(UNSET)
	header := frame.Header()
	debug("Send", header.TypeString(), header.FlagsString(), header.StreamId)
	err = frame.Write(transport.Conn)
	if err != nil {
		return err
	}

	return nil
}

func (transport *Transport) NewStream() (stream *Stream) {
	stream = NewStream(transport.NextId, transport.WriteChan)
	transport.Streams[transport.NextId] = stream
	transport.NextId += 2
	return stream
}

func (transport *Transport) ReadFrame() (frame Frame, err error) {
	header := new(FrameHeader)
	err = binary.Read(transport.Conn, binary.BigEndian, header)
	if err != nil {
		return nil, err
	}

	switch {
	case header.Type == SettingsFrameType:
		frame = &SettingsFrame{
			FrameHeader: header,
		}
	case header.Type == HeadersFrameType:
		length, _ := header.GetLength()
		b := make([]byte, length)
		err = binary.Read(transport.Conn, binary.BigEndian, b)
		if err != nil {
			return nil, err
		}
		headers, _ := Decode(b)
		frame = &HeadersFrame{
			FrameHeader: header,
			Headers:     headers,
		}
	case header.Type == DataFrameType:
		length, _ := header.GetLength()
		data := make([]byte, length)
		err = binary.Read(transport.Conn, binary.BigEndian, &data)
		frame = &DataFrame{
			FrameHeader: header,
			Data:        data,
		}
	case header.Type == GoAwayFrameType:
		var lastStreamId uint32
		err = binary.Read(transport.Conn, binary.BigEndian, &lastStreamId)
		var errorCode uint32
		err = binary.Read(transport.Conn, binary.BigEndian, &errorCode)
		frame = NewGoAwayFrame(lastStreamId, errorCode)
		go transport.Close()
	default:
		debug("unkown frame type")
	}

	return frame, nil
}

func (transport *Transport) ReadLoop() {
	for {
		frame, err := transport.ReadFrame()
		if err != nil {
			if err == io.EOF {
				debug("Got EOF")
				break
			}
			debug(err.Error())
			continue
		}
		streamId := frame.Header().StreamId

		stream, ok := transport.Streams[streamId]
		if !ok {
			debug("Create new stream", streamId)
			stream = NewStream(streamId, transport.WriteChan)
			transport.Streams[streamId] = stream
		}
		stream.ReadChan <- frame
	}
}

func (transport *Transport) ServerReadLoop(mux *http.ServeMux) {
	headerChan := make(chan http.Header)

	go func() {
		header := <-headerChan

		authority := header.Get(":authority")
		header.Del(":authority")
		method := header.Get(":method")
		header.Del(":method")
		path := header.Get(":path")
		header.Del(":path")
		scheme := header.Get(":scheme")
		header.Del(":scheme")

		rawurl := fmt.Sprintf("%s://%s%s", scheme, authority, path)
		reqUrl, err := url.ParseRequestURI(rawurl)

		if err != nil {
			debug(err)
		}

		req := &http.Request{
			Method:           method,
			URL:              reqUrl,
			Proto:            "HTTP/1.1",
			ProtoMajor:       1,
			ProtoMinor:       1,
			Header:           header,
			Body:             nil,
			ContentLength:    int64(0),
			TransferEncoding: []string{},
			Close:            false,
			Host:             authority,
		}

		rw := NewResponseWriter()
		mux.ServeHTTP(rw, req)
		responseHeader := rw.Header()
		responseHeader.Add(":status", strconv.Itoa(rw.Status))

		streamId := uint32(0x1)

		headersFrame := NewHeadersFrame(END_HEADERS, streamId)
		headersFrame.Headers = responseHeader
		transport.WriteFrame(headersFrame)

		dataFrame := NewDataFrame(END_STREAM, streamId, rw.Body.Bytes())
		transport.WriteFrame(dataFrame)
	}()

	for {
		frame, err := transport.ReadFrame()
		if err != nil {
			if err == io.EOF {
				debug("Got EOF")
				break
			}
			debug(err.Error())
			continue
		}
		streamId := frame.Header().StreamId

		stream, ok := transport.Streams[streamId]
		if !ok {
			debug("create new stream", streamId)
			stream = NewStream(streamId, transport.WriteChan)
			stream.HeaderChan = headerChan
			transport.Streams[streamId] = stream
		}
		stream.ReadChan <- frame
	}
}

func (transport *Transport) WriteFrame(frame Frame) (err error) {
	header := frame.Header()
	debug("Send", header.TypeString(), header.FlagsString(), header.StreamId)
	buf := new(bytes.Buffer)
	err = frame.Write(buf)
	if err != nil {
		return err
	}

	_, err = transport.Conn.Write(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (transport *Transport) WriteLoop() {
	for {
		select {
		case frame := <-transport.WriteChan:
			transport.WriteFrame(frame)
		}
	}
}

package minimum_http2

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
)

type Transport struct {
	Conn      net.Conn
	WriteChan chan Frame
	NextId    uint32
	Streams   map[uint32]*Stream
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

	go transport.ReadLoop()
	go transport.WriteLoop()

	transport.Streams[0] = NewStream(0, transport.WriteChan)

	transport.WriteConnectionPreface()
	transport.WriteEmptySettings()

	// wait server settings frame and settings frame ACK

	return nil
}

func (transport *Transport) NewStream() (stream *Stream) {
	stream = NewStream(transport.NextId, transport.WriteChan)
	transport.Streams[transport.NextId] = stream
	transport.NextId++
	return stream
}

func (transport *Transport) WriteEmptySettings() (err error) {
	frame := NewSettingsFrame(UNSET)
	err = transport.WriteFrame(frame)
	if err != nil {
		return err
	}

	return nil
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
		headers := map[string]string{}
		length, _ := header.GetLength()
		for i := 0; i < length; {
			b := make([]byte, 1)
			if i != 0 {
				return nil, errors.New("unkown header field representations")
			}
			i += 1
			err = binary.Read(transport.Conn, binary.BigEndian, &b)
			var nameLength uint8
			i += 1
			err = binary.Read(transport.Conn, binary.BigEndian, &nameLength)
			nameString := make([]byte, nameLength)
			i += int(nameLength)
			err = binary.Read(transport.Conn, binary.BigEndian, &nameString)
			var valueLength uint8
			i += 1
			err = binary.Read(transport.Conn, binary.BigEndian, &valueLength)
			valueString := make([]byte, valueLength)
			i += int(valueLength)
			err = binary.Read(transport.Conn, binary.BigEndian, &valueString)
			headers[string(nameString)] = string(valueString)
		}
		frame = &HeadersFrame{
			FrameHeader: header,
			Headers:     headers,
		}
	case header.Type == DataFrameType:
		length, _ := header.GetLength()
		data := make([]byte, length-1)
		err = binary.Read(transport.Conn, binary.BigEndian, &data)
		frame = &DataFrame{
			FrameHeader: header,
			Data:        data,
		}
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
			stream = NewStream(streamId, transport.WriteChan)
			transport.Streams[streamId] = stream
		}
		stream.ReadChan <- frame
	}
}

func (transport *Transport) WriteFrame(frame Frame) (err error) {
	header := frame.Header()
	debug("Send", header.TypeString(), header.FlagsString())
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

func (transport *Transport) WriteConnectionPreface() (err error) {
	debug("Send connection preface")
	_, err = transport.Conn.Write([]byte(CONNECTION_PREFACE))
	if err != nil {
		return err
	}

	return nil
}

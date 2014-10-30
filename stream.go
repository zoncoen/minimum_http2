package minimum_http2

import (
	"net/http"
)

type Stream struct {
	Id        uint32
	ReadChan  chan Frame
	WriteChan chan Frame
	Header    http.Header
}

func NewStream(id uint32, writeChan chan Frame) *Stream {
	stream := &Stream{
		Id:        id,
		ReadChan:  make(chan Frame),
		WriteChan: writeChan,
		Header:    http.Header{},
	}
	go stream.ReadLoop()
	return stream
}

func (stream *Stream) Write(frame Frame) {
	header := frame.Header()
	header.StreamId = stream.Id
	stream.WriteChan <- frame
}

func (stream *Stream) ReadLoop() (err error) {
	for {
		select {
		case frame := <-stream.ReadChan:
			header := frame.Header()
			if header.Type == SettingsFrameType {
				if header.Flags == UNSET {
					debug("Recv", header.TypeString(), header.FlagsString())
					ackFrame := NewSettingsFrame(ACK)
					stream.Write(ackFrame)
				}
				if header.Flags == ACK {
					debug("Recv", header.TypeString(), header.FlagsString())
				}
			}
			if header.Type == HeadersFrameType {
				debug("Recv", header.TypeString(), header.FlagsString())
				if f, ok := frame.(HeadersMaper); ok {
					for k, v := range f.HeadersMap() {
						stream.Header.Add(k, v)
					}
				}
			}
			if header.Type == DataFrameType {
				debug("Recv", header.TypeString(), header.FlagsString())
				if f, ok := frame.(DataByter); ok {
					debug(string(f.DataByte()))
				}
			}
		}
	}
}

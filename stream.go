package minimum_http2

import (
	"net/http"
)

type Stream struct {
	Id         uint32
	ReadChan   chan Frame
	WriteChan  chan Frame
	Header     http.Header
	Body       []byte
	HeaderChan chan http.Header
	BodyChan   chan []byte
	Quit       chan struct{}
}

func NewStream(id uint32, writeChan chan Frame) *Stream {
	stream := &Stream{
		Id:         id,
		ReadChan:   make(chan Frame),
		WriteChan:  writeChan,
		Header:     http.Header{},
		Body:       make([]byte, 0),
		HeaderChan: make(chan http.Header),
		BodyChan:   make(chan []byte),
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
					debug("Recv", header.TypeString(), header.FlagsString(), header.StreamId)
					ackFrame := NewSettingsFrame(ACK)
					stream.Write(ackFrame)
				}
				if header.Flags == ACK {
					debug("Recv", header.TypeString(), header.FlagsString(), header.StreamId)
				}
			}
			if header.Type == HeadersFrameType {
				debug("Recv", header.TypeString(), header.FlagsString(), header.StreamId)
				if f, ok := frame.(HeadersMapper); ok {
					for k, v := range f.HeadersMap() {
						stream.Header.Add(k, v[0])
					}
				}
				if header.Flags == END_HEADERS {
					stream.HeaderChan <- stream.Header
				}
			}
			if header.Type == DataFrameType {
				debug("Recv", header.TypeString(), header.FlagsString(), header.StreamId)
				if f, ok := frame.(DataByter); ok {
					stream.Body = append(stream.Body, f.DataByte()...)
				}
				if header.Flags == END_STREAM {
					stream.BodyChan <- stream.Body
				}
			}
			if header.Type == GoAwayFrameType {
				debug("Recv", header.TypeString(), header.FlagsString(), header.StreamId)
				break
			}
		case <-stream.Quit:
			break
		}
	}
}

package minimum_http2

import (
	"bytes"
	"encoding/binary"
	"io"
)

type Frame interface {
	Write(io.Writer) error
	Header() *FrameHeader
}

type SettingsFrame struct {
	FrameHeader *FrameHeader
}

func NewSettingsFrame(flags uint8) *SettingsFrame {
	header := &FrameHeader{
		Type:     SettingsFrameType,
		Flags:    flags,
		StreamId: 0x0,
	}
	header.SetLength(0x0)

	frame := &SettingsFrame{
		FrameHeader: header,
	}

	return frame
}

func (frame *SettingsFrame) Write(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, frame.FrameHeader)
	if err != nil {
		return err
	}
	return nil
}

func (frame *SettingsFrame) Header() *FrameHeader {
	return frame.FrameHeader
}

type HeadersMaper interface {
	HeadersMap() map[string]string
}

type HeadersFrame struct {
	FrameHeader      *FrameHeader
	PadLength        uint8
	StreamDependency uint32
	Weight           uint8
	Headers          map[string]string
}

func NewHeadersFrame(flags uint8, streamId uint32) *HeadersFrame {
	header := &FrameHeader{
		Type:     HeadersFrameType,
		Flags:    flags,
		StreamId: streamId,
	}

	frame := &HeadersFrame{
		FrameHeader: header,
		Headers:     map[string]string{},
	}

	return frame
}

func (frame *HeadersFrame) Header() *FrameHeader {
	return frame.FrameHeader
}

func (frame *HeadersFrame) AddHeader(name string, value string) {
	frame.Headers[name] = value
}

func (frame *HeadersFrame) Write(w io.Writer) (err error) {
	buf := new(bytes.Buffer)
	// Hpack
	packedHeaders, err := frame.Encode()
	if err != nil {
		return err
	}

	// set flags and length
	frame.FrameHeader.Flags = END_HEADERS
	length := 8 + 32 + 8 + len(packedHeaders)
	frame.FrameHeader.SetLength(length)

	// convert to bytes
	err = binary.Write(buf, binary.BigEndian, frame.FrameHeader)
	if err != nil {
		return err
	}
	err = binary.Write(buf, binary.BigEndian, frame.PadLength)
	if err != nil {
		return err
	}
	err = binary.Write(buf, binary.BigEndian, frame.StreamDependency)
	if err != nil {
		return err
	}
	err = binary.Write(buf, binary.BigEndian, frame.Weight)
	if err != nil {
		return err
	}
	_, err = buf.Write(packedHeaders)
	if err != nil {
		return err
	}

	// write
	_, err = w.Write(buf.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func (frame *HeadersFrame) HeadersMap() map[string]string {
	return frame.Headers
}

type DataByter interface {
	DataByte() []byte
}

type DataFrame struct {
	FrameHeader *FrameHeader
	Data        []byte
}

func (frame *DataFrame) Header() *FrameHeader {
	return frame.FrameHeader
}

func (frame *DataFrame) Write(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, frame.FrameHeader)
	if err != nil {
		return err
	}
	return nil
}

func (frame *DataFrame) DataByte() []byte {
	return frame.Data
}

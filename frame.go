package minimum_http2

import (
	"bytes"
	"encoding/binary"
	"io"
	"net/http"
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

type HeadersMapper interface {
	HeadersMap() http.Header
}

type HeadersFrame struct {
	FrameHeader *FrameHeader
	Headers     http.Header
}

func NewHeadersFrame(flags uint8, streamId uint32) *HeadersFrame {
	header := &FrameHeader{
		Type:     HeadersFrameType,
		Flags:    flags,
		StreamId: streamId,
	}

	frame := &HeadersFrame{
		FrameHeader: header,
		Headers:     make(http.Header),
	}

	return frame
}

func (frame *HeadersFrame) Header() *FrameHeader {
	return frame.FrameHeader
}

func (frame *HeadersFrame) AddHeader(name string, value string) {
	frame.Headers.Add(name, value)
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
	length := len(packedHeaders)
	frame.FrameHeader.SetLength(length)

	// convert to bytes
	err = binary.Write(buf, binary.BigEndian, frame.FrameHeader)
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

func (frame *HeadersFrame) HeadersMap() http.Header {
	return frame.Headers
}

type DataByter interface {
	DataByte() []byte
}

type DataFrame struct {
	FrameHeader *FrameHeader
	Data        []byte
}

func NewDataFrame(flags uint8, streamId uint32, data []byte) *DataFrame {
	header := &FrameHeader{
		Type:     DataFrameType,
		Flags:    flags,
		StreamId: streamId,
	}
	header.SetLength(len(data))

	frame := &DataFrame{
		FrameHeader: header,
		Data:        data,
	}

	return frame
}

func (frame *DataFrame) Header() *FrameHeader {
	return frame.FrameHeader
}

func (frame *DataFrame) Write(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, frame.FrameHeader)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, frame.Data)
	if err != nil {
		return err
	}
	return nil
}

func (frame *DataFrame) DataByte() []byte {
	return frame.Data
}

type GoAwayFrame struct {
	FrameHeader  *FrameHeader
	LastStreamId uint32
	ErrorCode    uint32
}

func NewGoAwayFrame(lastStreamId uint32, errorCode uint32) *GoAwayFrame {
	header := &FrameHeader{
		Type:     GoAwayFrameType,
		Flags:    UNSET,
		StreamId: uint32(0),
	}
	header.SetLength(64)

	frame := &GoAwayFrame{
		FrameHeader:  header,
		LastStreamId: lastStreamId,
		ErrorCode:    errorCode,
	}

	return frame
}

func (frame *GoAwayFrame) Write(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, frame.FrameHeader)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, frame.LastStreamId)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, frame.ErrorCode)
	if err != nil {
		return err
	}
	return nil
}

func (frame *GoAwayFrame) Header() *FrameHeader {
	return frame.FrameHeader
}

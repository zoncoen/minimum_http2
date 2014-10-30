package minimum_http2

import (
	"bytes"
	"encoding/binary"
)

// Types
const (
	DataFrameType uint8 = iota
	HeadersFrameType
	PriorityFrameType
	RstStreamFrameType
	SettingsFrameType
	PushPromiseFrameType
	PingFrameType
	GoAwayFrameType
	WindowUpdateFrameType
	ContinuationFrameType
)

// Flags
const (
	UNSET       uint8 = 0x0
	END_STREAM        = 0x1
	ACK               = 0x1
	END_HEADERS       = 0x4
	PADDED            = 0x8
	PRIORITY          = 0x20
)

type FrameHeader struct {
	Length   [3]byte
	Type     uint8
	Flags    uint8
	StreamId uint32
}

func (header *FrameHeader) SetLength(l int) (err error) {
	length := uint32(l)
	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, length)
	if err != nil {
		return err
	}
	array := buf.Bytes()
	header.Length = [3]byte{array[1], array[2], array[3]}
	return nil
}

func (header *FrameHeader) GetLength() (int, error) {
	b := []byte{0x0, header.Length[0], header.Length[1], header.Length[2]}
	buf := bytes.NewBuffer(b)
	var length uint32
	err := binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		return 0, err
	}
	return int(length), err
}

func (header *FrameHeader) TypeString() string {
	typeStrings := []string{
		"DataFrame",
		"HeadersFrame",
		"PriorityFrame",
		"RstStreamFrame",
		"SettingsFrame",
		"PushPromiseFrame",
		"PingFrame",
		"GoAwayFrame",
		"WindowUpdateFrame",
		"ContinuationFrame",
	}
	return typeStrings[header.Type]
}

func (header *FrameHeader) FlagsString() string {
	flagsStrings := make([]string, 0x20+1)
	flagsStrings[UNSET] = "UNSET"
	flagsStrings[END_STREAM] = "END_STREAM"
	flagsStrings[END_HEADERS] = "END_HEADERS"
	flagsStrings[PADDED] = "PADDED"
	flagsStrings[PRIORITY] = "PRIORITY"

	if header.Type == SettingsFrameType {
		flagsStrings[ACK] = "ACK"
	}

	return flagsStrings[header.Flags]
}

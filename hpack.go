package minimum_http2

import (
	"bytes"
	"encoding/binary"
)

// Literal Header Field without Indexing
func (frame *HeadersFrame) Encode() ([]byte, error) {
	// TODO: handle long length strings (Integer Representation)
	buf := new(bytes.Buffer)
	var pad, length uint8 = 0, 0
	for k, v := range frame.Headers {
		binary.Write(buf, binary.BigEndian, pad)
		length = uint8(len(k))
		binary.Write(buf, binary.BigEndian, length)
		binary.Write(buf, binary.BigEndian, k)
		length = uint8(len(v))
		binary.Write(buf, binary.BigEndian, length)
		binary.Write(buf, binary.BigEndian, v)
	}

	return buf.Bytes(), nil
}

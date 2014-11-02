package minimum_http2

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net/http"
)

// Literal Header Field without Indexing
func (frame *HeadersFrame) Encode() ([]byte, error) {
	// TODO: handle long length strings (Integer Representation)
	buf := new(bytes.Buffer)
	var headerField, length uint8 = 0, 0
	for k, v := range frame.Headers {
		binary.Write(buf, binary.BigEndian, headerField)
		length = uint8(len(k))
		binary.Write(buf, binary.BigEndian, length)
		binary.Write(buf, binary.BigEndian, []byte(k))
		length = uint8(len(v[0]))
		binary.Write(buf, binary.BigEndian, length)
		binary.Write(buf, binary.BigEndian, []byte(v[0]))
	}

	return buf.Bytes(), nil
}

// Literal Header Field without Indexing
func Decode(b []byte) (http.Header, error) {
	buf := bytes.NewBuffer(b)
	headers := make(http.Header)
	for i := 0; i < len(b); {
		// detect format of header field representations
		var headerField uint8
		err := binary.Read(buf, binary.BigEndian, &headerField)
		if headerField != 0 {
			return nil, errors.New("unkown header field (this library implements 'Indexed Header Field Representation' only)")
		}
		i += 1

		var nameLength uint8
		err = binary.Read(buf, binary.BigEndian, &nameLength)
		if err != nil {
			return nil, err
		}
		i += 1

		nameString := make([]byte, nameLength)
		err = binary.Read(buf, binary.BigEndian, &nameString)
		if err != nil {
			return nil, err
		}
		i += int(nameLength)

		var valueLength uint8
		err = binary.Read(buf, binary.BigEndian, &valueLength)
		if err != nil {
			return nil, err
		}
		i += 1

		valueString := make([]byte, valueLength)
		err = binary.Read(buf, binary.BigEndian, &valueString)
		if err != nil {
			return nil, err
		}
		i += int(valueLength)

		headers.Add(string(nameString), string(valueString))
	}
	return headers, nil
}

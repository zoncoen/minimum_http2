package minimum_http2

import ()

type Client struct {
	Transport *Transport
}

func (client *Client) Connect(addr string) error {
	client.Transport = &Transport{}
	err := client.Transport.Connect(addr)
	if err != nil {
		return err
	}
	return nil
}

func (client *Client) Disconnect() {
	frame := NewGoAwayFrame(client.Transport.NextId-2, uint32(0))
	frame.Write(client.Transport.Conn)
}

func (client *Client) Get(path string) *Response {
	stream := client.Transport.NewStream()
	frame := NewHeadersFrame(UNSET, 0x1)
	frame.AddHeader(":method", "GET")
	frame.AddHeader(":scheme", "http")
	frame.AddHeader(":authority", client.Transport.Conn.RemoteAddr().String())
	frame.AddHeader(":path", path)
	stream.WriteChan <- frame
	header := <-stream.HeaderChan
	body := <-stream.BodyChan
	return &Response{
		Header: header,
		Body:   body,
	}
}

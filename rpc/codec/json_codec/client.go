package json_codec

import (
	"encoding/json"
	"io"
	"sync"

	"github.com/ironzhang/zerone/rpc/codec"
)

var _ codec.ClientCodec = &ClientCodec{}

type ClientCodec struct {
	mu   sync.Mutex
	rwc  io.ReadWriteCloser
	enc  *json.Encoder
	dec  *json.Decoder
	req  clientRequest
	resp clientResponse
}

func NewClientCodec(rwc io.ReadWriteCloser) *ClientCodec {
	return &ClientCodec{
		rwc: rwc,
		enc: json.NewEncoder(rwc),
		dec: json.NewDecoder(rwc),
	}
}

func (c *ClientCodec) WriteRequest(h *codec.RequestHeader, x interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.req.ClassMethod = h.ClassMethod
	c.req.Sequence = h.Sequence
	c.req.TraceID = h.TraceID
	c.req.ClientName = h.ClientName
	c.req.Verbose = h.Verbose
	c.req.Body = x
	return c.enc.Encode(&c.req)
}

func (c *ClientCodec) reset() {
	c.resp.ClassMethod = ""
	c.resp.Sequence = 0
	c.resp.Code = 0
	c.resp.Cause = ""
	c.resp.Desc = ""
	c.resp.ServerName = ""
	c.resp.Body = nil
}

func (c *ClientCodec) ReadResponseHeader(h *codec.ResponseHeader) error {
	c.reset()

	if err := c.dec.Decode(&c.resp); err != nil {
		return err
	}

	h.ClassMethod = c.resp.ClassMethod
	h.Sequence = c.resp.Sequence
	h.Error.Code = c.resp.Code
	h.Error.Cause = c.resp.Cause
	h.Error.Desc = c.resp.Desc
	h.Error.ServerName = c.resp.ServerName
	return nil
}

func (c *ClientCodec) ReadResponseBody(x interface{}) error {
	if x == nil {
		return nil
	}
	return json.Unmarshal(c.resp.Body, x)
}

func (c *ClientCodec) Close() error {
	return c.rwc.Close()
}

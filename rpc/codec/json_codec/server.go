package json_codec

import (
	"encoding/json"
	"io"
	"sync"

	"github.com/ironzhang/zerone/rpc/codec"
)

var _ codec.ServerCodec = &ServerCodec{}

type ServerCodec struct {
	mu   sync.Mutex
	rwc  io.ReadWriteCloser
	enc  *json.Encoder
	dec  *json.Decoder
	req  serverRequest
	resp serverResponse
}

func NewServerCodec(rwc io.ReadWriteCloser) *ServerCodec {
	return &ServerCodec{
		rwc: rwc,
		enc: json.NewEncoder(rwc),
		dec: json.NewDecoder(rwc),
	}
}

func (c *ServerCodec) reset() {
	c.req.ClassMethod = ""
	c.req.Sequence = 0
	c.req.TraceID = ""
	c.req.ClientName = ""
	c.req.Verbose = 0
	c.req.Body = nil
}

func (c *ServerCodec) ReadRequestHeader(h *codec.RequestHeader) error {
	c.reset()

	if err := c.dec.Decode(&c.req); err != nil {
		return err
	}

	h.ClassMethod = c.req.ClassMethod
	h.Sequence = c.req.Sequence
	h.TraceID = c.req.TraceID
	h.ClientName = c.req.ClientName
	h.Verbose = c.req.Verbose
	return nil
}

func (c *ServerCodec) ReadRequestBody(x interface{}) error {
	if x == nil {
		return nil
	}
	return json.Unmarshal(c.req.Body, x)
}

func (c *ServerCodec) WriteResponse(h *codec.ResponseHeader, x interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.resp.ClassMethod = h.ClassMethod
	c.resp.Sequence = h.Sequence
	c.resp.Code = h.Error.Code
	c.resp.Cause = h.Error.Cause
	c.resp.Desc = h.Error.Desc
	c.resp.ServerName = h.Error.ServerName
	c.resp.Body = x
	return c.enc.Encode(&c.resp)
}

func (c *ServerCodec) Close() error {
	return c.rwc.Close()
}

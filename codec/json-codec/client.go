package json_codec

import (
	"encoding/json"
	"io"

	"github.com/ironzhang/zerone/codec"
)

type clientRequest struct {
	Method   string
	Sequence uint64
	TraceID  string
	ClientID string
	Verbose  bool
	Body     interface{} `json:",omitempty"`
}

type clientResponse struct {
	Method   string
	Sequence uint64
	Code     int
	Desc     string
	Cause    string
	Body     json.RawMessage
}

type clientCodec struct {
	enc  *json.Encoder
	dec  *json.Decoder
	req  clientRequest
	resp clientResponse
}

func NewClientCodec(rw io.ReadWriter) codec.ClientCodec {
	return &clientCodec{
		enc: json.NewEncoder(rw),
		dec: json.NewDecoder(rw),
	}
}

func (c *clientCodec) WriteRequest(h *codec.RequestHeader, x interface{}) error {
	c.req.Method = h.Method
	c.req.Sequence = h.Sequence
	c.req.TraceID = h.TraceID
	c.req.ClientID = h.ClientID
	c.req.Verbose = h.Verbose
	c.req.Body = x
	return c.enc.Encode(&c.req)
}

func (c *clientCodec) ReadResponseHeader(h *codec.ResponseHeader) error {
	if err := c.dec.Decode(&c.resp); err != nil {
		return err
	}

	h.Method = c.resp.Method
	h.Sequence = c.resp.Sequence
	h.Code = c.resp.Code
	h.Desc = c.resp.Desc
	h.Cause = c.resp.Cause
	return nil
}

func (c *clientCodec) ReadResponseBody(x interface{}) error {
	if x == nil {
		return nil
	}
	return json.Unmarshal(c.resp.Body, x)
}

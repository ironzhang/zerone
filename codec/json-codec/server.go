package json_codec

import (
	"encoding/json"
	"io"

	"github.com/ironzhang/zerone/codec"
)

type serverRequest struct {
	Method   string
	Sequence uint64
	TraceID  string
	ClientID string
	Verbose  bool
	Body     json.RawMessage
}

type serverResponse struct {
	Method   string
	Sequence uint64
	Code     int
	Desc     string      `json:",omitempty"`
	Cause    string      `json:",omitempty"`
	Body     interface{} `json:",omitempty"`
}

type serverCodec struct {
	enc  *json.Encoder
	dec  *json.Decoder
	req  serverRequest
	resp serverResponse
}

func NewServerCodec(rw io.ReadWriter) codec.ServerCodec {
	return &serverCodec{
		enc: json.NewEncoder(rw),
		dec: json.NewDecoder(rw),
	}
}

func (c *serverCodec) ReadRequestHeader(h *codec.RequestHeader) error {
	if err := c.dec.Decode(&c.req); err != nil {
		return err
	}

	h.Method = c.req.Method
	h.Sequence = c.req.Sequence
	h.TraceID = c.req.TraceID
	h.ClientID = c.req.ClientID
	h.Verbose = c.req.Verbose
	return nil
}

func (c *serverCodec) ReadRequestBody(x interface{}) error {
	if x == nil {
		return nil
	}
	return json.Unmarshal(c.req.Body, x)
}

func (c *serverCodec) WriteResponse(h *codec.ResponseHeader, x interface{}) error {
	c.resp.Method = h.Method
	c.resp.Sequence = h.Sequence
	c.resp.Code = h.Code
	c.resp.Desc = h.Desc
	c.resp.Cause = h.Cause
	c.resp.Body = x
	return c.enc.Encode(&c.resp)
}

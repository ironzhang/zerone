package json_codec

import (
	"encoding/json"
	"io"

	"github.com/ironzhang/zerone/rpc/codec"
)

type clientRequest struct {
	ServiceMethod string      `json:"ServiceMethod"`
	Sequence      uint64      `json:"Sequence"`
	TraceID       string      `json:"TraceID"`
	ClientName    string      `json:"ClientName"`
	Verbose       bool        `json:"Verbose,omitempty"`
	Cancel        bool        `json:"Cancel,omitempty"`
	Body          interface{} `json:"Body,omitempty"`
}

type clientResponse struct {
	ServiceMethod string          `json:"ServiceMethod"`
	Sequence      uint64          `json:"Sequence"`
	Code          int             `json:"Code"`
	Desc          string          `json:"Desc,omitempty"`
	Cause         string          `json:"Cause,omitempty"`
	ServerName    string          `json:"ServerName,omitempty"`
	Body          json.RawMessage `json:"Body,omitempty"`
}

var _ codec.ClientCodec = &ClientCodec{}

type ClientCodec struct {
	enc  *json.Encoder
	dec  *json.Decoder
	req  clientRequest
	resp clientResponse
}

func NewClientCodec(rw io.ReadWriter) *ClientCodec {
	return &ClientCodec{
		enc: json.NewEncoder(rw),
		dec: json.NewDecoder(rw),
	}
}

func (c *ClientCodec) WriteRequest(h *codec.RequestHeader, x interface{}) error {
	c.req.ServiceMethod = h.ServiceMethod
	c.req.Sequence = h.Sequence
	c.req.TraceID = h.TraceID
	c.req.ClientName = h.ClientName
	c.req.Verbose = h.Verbose
	c.req.Cancel = h.Cancel
	c.req.Body = x
	return c.enc.Encode(&c.req)
}

func (c *ClientCodec) reset() {
	c.resp.ServiceMethod = ""
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

	h.ServiceMethod = c.resp.ServiceMethod
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

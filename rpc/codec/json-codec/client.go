package json_codec

import (
	"encoding/json"
	"io"

	"github.com/ironzhang/zerone/rpc/codec"
)

type clientRequest struct {
	Method     string      `json:"Method"`
	Sequence   uint64      `json:"Sequence"`
	TraceID    string      `json:"TraceID"`
	ClientName string      `json:"ClientName"`
	Verbose    bool        `json:"Verbose"`
	Body       interface{} `json:"Body,omitempty"`
}

type clientResponse struct {
	Method      string          `json:"Method"`
	Sequence    uint64          `json:"Sequence"`
	Code        int             `json:"Code"`
	Message     string          `json:"Message,omitempty"`
	Description string          `json:"Description,omitempty"`
	ServerName  string          `json:"ServerName,omitempty"`
	Body        json.RawMessage `json:"Body,omitempty"`
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
	c.req.ClientName = h.ClientName
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
	h.Error.Code = c.resp.Code
	h.Error.Message = c.resp.Message
	h.Error.Description = c.resp.Description
	h.Error.ServerName = c.resp.ServerName
	return nil
}

func (c *clientCodec) ReadResponseBody(x interface{}) error {
	if x == nil {
		return nil
	}
	return json.Unmarshal(c.resp.Body, x)
}

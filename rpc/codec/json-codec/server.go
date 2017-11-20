package json_codec

import (
	"encoding/json"
	"io"

	"github.com/ironzhang/zerone/rpc/codec"
)

type serverRequest struct {
	Method     string          `json:"Method"`
	Sequence   uint64          `json:"Sequence"`
	TraceID    string          `json:"TraceID"`
	ClientName string          `json:"ClientName"`
	Verbose    bool            `json:"Verbose"`
	Body       json.RawMessage `json:"Body,omitempty"`
}

type serverResponse struct {
	Method      string      `json:"Method"`
	Sequence    uint64      `json:"Sequence"`
	Code        int         `json:"Code"`
	Message     string      `json:"Message,omitempty"`
	Description string      `json:"Description,omitempty"`
	ServerName  string      `json:"ServerName,omitempty"`
	Body        interface{} `json:"Body,omitempty"`
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
	h.ClientName = c.req.ClientName
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
	c.resp.Code = h.Error.Code
	c.resp.Message = h.Error.Message
	c.resp.Description = h.Error.Description
	c.resp.ServerName = h.Error.ServerName
	c.resp.Body = x
	return c.enc.Encode(&c.resp)
}

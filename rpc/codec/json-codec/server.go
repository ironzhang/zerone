package json_codec

import (
	"encoding/json"
	"io"

	"github.com/ironzhang/zerone/rpc/codec"
)

type serverRequest struct {
	ServiceMethod string          `json:"ServiceMethod"`
	Sequence      uint64          `json:"Sequence"`
	TraceID       string          `json:"TraceID"`
	ClientName    string          `json:"ClientName"`
	Verbose       bool            `json:"Verbose,omitempty"`
	Cancel        bool            `json:"Cancel,omitempty"`
	Body          json.RawMessage `json:"Body,omitempty"`
}

type serverResponse struct {
	ServiceMethod string      `json:"ServiceMethod"`
	Sequence      uint64      `json:"Sequence"`
	Code          int         `json:"Code"`
	Message       string      `json:"Message,omitempty"`
	Description   string      `json:"Description,omitempty"`
	ServerName    string      `json:"ServerName,omitempty"`
	Body          interface{} `json:"Body,omitempty"`
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

func (c *serverCodec) reset() {
	c.req.ServiceMethod = ""
	c.req.Sequence = 0
	c.req.TraceID = ""
	c.req.ClientName = ""
	c.req.Verbose = false
	c.req.Cancel = false
	c.req.Body = nil
}

func (c *serverCodec) ReadRequestHeader(h *codec.RequestHeader) error {
	c.reset()

	if err := c.dec.Decode(&c.req); err != nil {
		return err
	}

	h.ServiceMethod = c.req.ServiceMethod
	h.Sequence = c.req.Sequence
	h.TraceID = c.req.TraceID
	h.ClientName = c.req.ClientName
	h.Verbose = c.req.Verbose
	h.Cancel = c.req.Cancel
	return nil
}

func (c *serverCodec) ReadRequestBody(x interface{}) error {
	if x == nil {
		return nil
	}
	return json.Unmarshal(c.req.Body, x)
}

func (c *serverCodec) WriteResponse(h *codec.ResponseHeader, x interface{}) error {
	c.resp.ServiceMethod = h.ServiceMethod
	c.resp.Sequence = h.Sequence
	c.resp.Code = h.Error.Code
	c.resp.Message = h.Error.Message
	c.resp.Description = h.Error.Description
	c.resp.ServerName = h.Error.ServerName
	c.resp.Body = x
	return c.enc.Encode(&c.resp)
}

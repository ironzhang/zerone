package json_codec

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"os"
	"reflect"
	"testing"

	"github.com/ironzhang/zerone/rpc/codec"
)

type Args struct {
	A, B int
}

type Reply struct {
	C int
}

func TestRequest(t *testing.T) {
	tests := []clientRequest{
		{
			ServiceMethod: "Add",
			Sequence:      1,
			TraceID:       "1",
			ClientName:    "client-1",
			Verbose:       true,
			Cancel:        false,
			Body:          &Args{A: 1, B: 2},
		},
		{
			ServiceMethod: "Ping",
			Sequence:      1,
			TraceID:       "1",
			ClientName:    "client-1",
			Verbose:       true,
			Cancel:        false,
			Body:          "hello",
		},
		{
			ServiceMethod: "Ping",
			Sequence:      1,
			TraceID:       "1",
			ClientName:    "client-1",
			Verbose:       true,
			Cancel:        true,
			Body:          nil,
		},
	}
	for i, creq := range tests {
		var sreq serverRequest
		data, err := json.Marshal(creq)
		if err != nil {
			t.Fatalf("case%d: json marshal: %v", i, err)
		}
		if err = json.Unmarshal(data, &sreq); err != nil {
			t.Fatalf("case%d: json unmarshal: %v", i, err)
		}
		body, err := json.Marshal(creq.Body)
		if err != nil {
			t.Fatalf("case%d: marshal body: %v", i, err)
		}

		if got, want := sreq.ServiceMethod, creq.ServiceMethod; got != want {
			t.Fatalf("case%d: ServiceMethod: %v != %v", i, got, want)
		}
		if got, want := sreq.Sequence, creq.Sequence; got != want {
			t.Fatalf("case%d: Sequence: %v != %v", i, got, want)
		}
		if got, want := sreq.TraceID, creq.TraceID; got != want {
			t.Fatalf("case%d: Sequence: %v != %v", i, got, want)
		}
		if got, want := sreq.ClientName, creq.ClientName; got != want {
			t.Fatalf("case%d: Sequence: %v != %v", i, got, want)
		}
		if got, want := sreq.Verbose, creq.Verbose; got != want {
			t.Fatalf("case%d: Sequence: %v != %v", i, got, want)
		}
		if got, want := sreq.Cancel, creq.Cancel; got != want {
			t.Fatalf("case%d: Sequence: %v != %v", i, got, want)
		}
		if creq.Body != nil {
			if got, want := sreq.Body, body; !bytes.Equal(got, want) {
				t.Fatalf("case%d: Body: %s != %s", i, got, want)
			}
		}
	}
}

func TestResponse(t *testing.T) {
	tests := []serverResponse{
		{
			ServiceMethod: "Add",
			Sequence:      1,
			Code:          0,
			Cause:         "",
			Desc:          "",
			Module:        "",
			Body:          &Reply{C: 3},
		},
		{
			ServiceMethod: "Add",
			Sequence:      1,
			Code:          1,
			Cause:         "Message",
			Desc:          "Description",
			Module:        "ServerName",
			Body:          nil,
		},
		{
			ServiceMethod: "Ping",
			Sequence:      2,
			Code:          0,
			Body:          "hello",
		},
	}
	for i, sresp := range tests {
		var cresp clientResponse
		data, err := json.Marshal(sresp)
		if err != nil {
			t.Fatalf("case%d: json marshal: %v", i, err)
		}
		if err = json.Unmarshal(data, &cresp); err != nil {
			t.Fatalf("case%d: json umarshal: %v", i, err)
		}
		body, err := json.Marshal(sresp.Body)
		if err != nil {
			t.Fatalf("case%d: json marshal body: %v", i, err)
		}

		if got, want := cresp.ServiceMethod, sresp.ServiceMethod; got != want {
			t.Fatalf("case%d: ServiceMethod: %v != %v", i, got, want)
		}
		if got, want := cresp.Sequence, sresp.Sequence; got != want {
			t.Fatalf("case%d: Sequence: %v != %v", i, got, want)
		}
		if got, want := cresp.Code, sresp.Code; got != want {
			t.Fatalf("case%d: Code: %v != %v", i, got, want)
		}
		if got, want := cresp.Cause, sresp.Cause; got != want {
			t.Fatalf("case%d: Message: %v != %v", i, got, want)
		}
		if got, want := cresp.Desc, sresp.Desc; got != want {
			t.Fatalf("case%d: Description: %v != %v", i, got, want)
		}
		if got, want := cresp.Module, sresp.Module; got != want {
			t.Fatalf("case%d: ServerName: %v != %v", i, got, want)
		}
		if sresp.Body != nil {
			if got, want := cresp.Body, body; !bytes.Equal(got, want) {
				t.Fatalf("case%d: Body: %s != %s", i, got, want)
			}
		}
	}
}

type CompositeReadWriters struct {
	master  io.ReadWriter
	writers []io.Writer
}

func NewCompositeReadWriters(m io.ReadWriter, ws ...io.Writer) *CompositeReadWriters {
	return &CompositeReadWriters{
		master:  m,
		writers: ws,
	}
}

func (p *CompositeReadWriters) Write(b []byte) (n int, err error) {
	for _, w := range p.writers {
		w.Write(b)
	}
	return p.master.Write(b)
}

func (p *CompositeReadWriters) Read(b []byte) (n int, err error) {
	return p.master.Read(b)
}

func (p *CompositeReadWriters) Close() error {
	return nil
}

func TestWriteReadRequest(t *testing.T) {
	cli, svr := net.Pipe()
	defer func() {
		cli.Close()
		svr.Close()
	}()

	c := NewClientCodec(NewCompositeReadWriters(cli, os.Stdout))
	s := NewServerCodec(svr)

	s1, s2 := "hello", ""
	tests := []struct {
		h codec.RequestHeader
		x interface{}
		y interface{}
	}{
		{
			h: codec.RequestHeader{
				ServiceMethod: "Add",
				Sequence:      1,
				TraceID:       "1",
				ClientName:    "client-1",
				Verbose:       true,
			},
			x: &Args{A: 1, B: 2},
			y: &Args{},
		},
		{
			h: codec.RequestHeader{
				ServiceMethod: "Add",
				Sequence:      1,
				TraceID:       "1",
				ClientName:    "client-1",
				Verbose:       true,
				Cancel:        true,
			},
			x: nil,
			y: nil,
		},
		{
			h: codec.RequestHeader{
				ServiceMethod: "Ping",
				Sequence:      2,
				TraceID:       "2",
				ClientName:    "client-2",
				Verbose:       false,
			},
			x: &s1,
			y: &s2,
		},
		{
			h: codec.RequestHeader{
				ServiceMethod: "Ping",
				Sequence:      3,
				TraceID:       "3",
				ClientName:    "client-3",
				Verbose:       false,
			},
			x: nil,
			y: nil,
		},
	}
	for i, tt := range tests {
		go func(h *codec.RequestHeader, x interface{}) {
			if err := c.WriteRequest(h, x); err != nil {
				t.Fatalf("write request: %v", err)
			}
		}(&tt.h, tt.x)

		var h codec.RequestHeader
		if err := s.ReadRequestHeader(&h); err != nil {
			t.Fatalf("read request header: %v", err)
		}
		if got, want := h, tt.h; got != want {
			t.Fatalf("case%d: header: %v != %v", i, got, want)
		}
		if err := s.ReadRequestBody(tt.y); err != nil {
			t.Fatalf("read request body: %v", err)
		}
		if got, want := tt.y, tt.x; !reflect.DeepEqual(got, want) {
			t.Fatalf("case%d: body: %v != %v", i, got, want)
		}
	}
}

func TestWriteReadResponse(t *testing.T) {
	cli, svr := net.Pipe()
	defer func() {
		cli.Close()
		svr.Close()
	}()

	c := NewClientCodec(cli)
	s := NewServerCodec(NewCompositeReadWriters(svr, os.Stdout))

	s1, s2 := "", "hello"
	tests := []struct {
		h codec.ResponseHeader
		x interface{}
		y interface{}
	}{
		{
			h: codec.ResponseHeader{
				ServiceMethod: "Add",
				Sequence:      1,
			},
			x: &Reply{C: 3},
			y: &Reply{},
		},
		{
			h: codec.ResponseHeader{
				ServiceMethod: "Add",
				Sequence:      1,
				Error: codec.Error{
					Code:   1,
					Cause:  "Message",
					Desc:   "Description",
					Module: "ServerName",
				},
			},
			x: nil,
			y: nil,
		},
		{
			h: codec.ResponseHeader{
				ServiceMethod: "Ping",
				Sequence:      2,
			},
			x: &s1,
			y: &s2,
		},
		{
			h: codec.ResponseHeader{
				ServiceMethod: "Ping",
				Sequence:      3,
				Error: codec.Error{
					Code:   1,
					Cause:  "Message",
					Desc:   "Description",
					Module: "ServerName",
				},
			},
			x: nil,
			y: nil,
		},
	}
	var h codec.ResponseHeader
	for i, tt := range tests {
		go func(h *codec.ResponseHeader, x interface{}) {
			if err := s.WriteResponse(h, x); err != nil {
				t.Fatalf("write response: %v", err)
			}
		}(&tt.h, tt.x)
		if err := c.ReadResponseHeader(&h); err != nil {
			t.Fatalf("read response header: %v", err)
		}
		if got, want := h, tt.h; got != want {
			t.Fatalf("case%d: header: %v != %v", i, got, want)
		}
		if err := c.ReadResponseBody(tt.y); err != nil {
			t.Fatalf("read response body: %v", err)
		}
		if got, want := tt.y, tt.x; !reflect.DeepEqual(got, want) {
			t.Fatalf("case%d: body: %v != %v", i, got, want)
		}
	}
}

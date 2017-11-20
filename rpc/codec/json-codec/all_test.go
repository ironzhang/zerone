package json_codec

import (
	"io"
	"net"
	"os"
	"reflect"
	"testing"

	"github.com/ironzhang/zerone/rpc/codec"
)

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

type Args struct {
	A, B int
}

type Reply struct {
	C int
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
				Method:   "Add",
				Sequence: 1,
				TraceID:  "1",
				ClientID: "1",
				Verbose:  true,
			},
			x: &Args{A: 1, B: 2},
			y: &Args{},
		},
		{
			h: codec.RequestHeader{
				Method:   "Ping",
				Sequence: 2,
				TraceID:  "2",
				ClientID: "2",
				Verbose:  false,
			},
			x: &s1,
			y: &s2,
		},
		{
			h: codec.RequestHeader{
				Method:   "Ping",
				Sequence: 3,
				TraceID:  "3",
				ClientID: "3",
				Verbose:  false,
			},
			x: nil,
			y: nil,
		},
	}
	var h codec.RequestHeader
	for i, tt := range tests {
		go func(h *codec.RequestHeader, x interface{}) {
			if err := c.WriteRequest(h, x); err != nil {
				t.Fatalf("write request: %v", err)
			}
		}(&tt.h, tt.x)

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
				Method:   "Add",
				Sequence: 1,
				Code:     0,
				Desc:     "",
				Cause:    "",
			},
			x: &Reply{C: 3},
			y: &Reply{},
		},
		{
			h: codec.ResponseHeader{
				Method:   "Ping",
				Sequence: 2,
				Code:     0,
				Desc:     "",
				Cause:    "",
			},
			x: &s1,
			y: &s2,
		},
		{
			h: codec.ResponseHeader{
				Method:   "Ping",
				Sequence: 3,
				Code:     1,
				Desc:     "error",
				Cause:    "error",
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

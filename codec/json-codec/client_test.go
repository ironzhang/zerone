package json_codec

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ironzhang/zerone/codec"
)

func TestClientCodecWriteRequest(t *testing.T) {
	requests := []struct {
		h codec.RequestHeader
		x interface{}
	}{
		{
			h: codec.RequestHeader{
				Method:   "Method-1",
				Sequence: 0,
				TraceID:  "TraceID-1",
				ClientID: "ClientID-1",
				Verbose:  true,
			},
			x: struct {
				A int
				B string
			}{
				A: 1,
				B: "hello, world",
			},
		},
		{
			h: codec.RequestHeader{
				Method:   "Method-2",
				Sequence: 1,
				TraceID:  "TraceID-2",
				ClientID: "ClientID-2",
				Verbose:  false,
			},
			x: struct {
				C float64
				D bool
			}{
				C: 1.0,
				D: true,
			},
		},
	}

	var want string
	want = want + `{"Method":"Method-1","Sequence":0,"TraceID":"TraceID-1","ClientID":"ClientID-1","Verbose":true,"Body":{"A":1,"B":"Hello, world"}}` + "\n"
	want = want + `{"Method":"Method-2","Sequence":1,"TraceID":"TraceID-2","ClientID":"ClientID-2","Verbose":false,"Body":{"C":1,"D":true}}` + "\n"

	var b bytes.Buffer
	c := NewClientCodec(&b)
	for _, req := range requests {
		if err := c.WriteRequest(&req.h, req.x); err != nil {
			t.Fatalf("write request: %v", err)
		}
	}
	if got := b.String(); !strings.EqualFold(got, want) {
		t.Errorf("%s != %s", got, want)
	}
}

func TestClientCodecReadResponse(t *testing.T) {
}

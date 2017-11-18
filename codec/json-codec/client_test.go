package json_codec

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/ironzhang/zerone/codec"
)

type Args struct {
	A, B int
}

type Reply struct {
	C int
}

type Ping struct {
	S string
}

type Pong struct {
	S string
}

var (
	reqstr = "" +
		`{"Method":"Add","Sequence":0,"TraceID":"TraceID-1","ClientID":"ClientID-1","Verbose":true,"Body":{"A":1,"B":2}}` + "\n" +
		`{"Method":"Ping","Sequence":1,"TraceID":"TraceID-2","ClientID":"ClientID-2","Verbose":false,"Body":{"S":"Ping"}}` + "\n"

	respstr = "" +
		`{"Method":"Add","Sequence":0,"Code":0,"Body":{"C":3}}` + "\n" +
		`{"Method":"Ping","Sequence":1,"Code":1,"Desc":"Desc","Cause":"Cause"}` + "\n"
)

func TestClientCodecWriteRequest(t *testing.T) {
	requests := []struct {
		h codec.RequestHeader
		x interface{}
	}{
		{
			h: codec.RequestHeader{
				Method: "Add", Sequence: 0, TraceID: "TraceID-1", ClientID: "ClientID-1", Verbose: true,
			},
			x: Args{A: 1, B: 2},
		},
		{
			h: codec.RequestHeader{
				Method: "Ping", Sequence: 1, TraceID: "TraceID-2", ClientID: "ClientID-2", Verbose: false,
			},
			x: Ping{S: "Ping"},
		},
	}

	var b bytes.Buffer
	c := NewClientCodec(&b)
	for _, req := range requests {
		if err := c.WriteRequest(&req.h, req.x); err != nil {
			t.Fatalf("write request: %v", err)
		}
	}
	if got, want := b.String(), reqstr; !strings.EqualFold(got, want) {
		t.Errorf("%s != %s", got, want)
	}
}

func TestClientCodecReadResponse(t *testing.T) {
	type response struct {
		h codec.ResponseHeader
		x interface{}
	}

	tests := []struct {
		got  response
		want response
	}{
		//`{"Method":"Add","Sequence":0,"Code":0,"Body":{"C":3}}` + "\n"
		{
			got: response{x: &Reply{}},
			want: response{
				h: codec.ResponseHeader{
					Method:   "Add",
					Sequence: 0,
					Code:     0,
				},
				x: &Reply{C: 3},
			},
		},
		//`{"Method":"Ping","Sequence":1,"Code":1,"Desc":"Desc","Cause":"Cause"}` + "\n"
		{
			got: response{x: &Pong{}},
			want: response{
				h: codec.ResponseHeader{
					Method:   "Ping",
					Sequence: 1,
					Code:     1,
					Desc:     "Desc",
					Cause:    "Cause",
				},
				x: &Pong{},
			},
		},
	}

	c := NewClientCodec(bytes.NewBufferString(respstr))
	for i, tt := range tests {
		if err := c.ReadResponseHeader(&tt.got.h); err != nil {
			t.Fatalf("read response header: %v", err)
		}
		if err := c.ReadResponseBody(&tt.got.x); err != nil {
			t.Fatalf("read response body: %v", err)
		}
		if got, want := tt.got.h, tt.want.h; got != want {
			t.Errorf("case%d: header: %v != %v", i, got, want)
		}
		if got, want := tt.got.x, tt.want.x; !reflect.DeepEqual(got, want) {
			t.Errorf("case%d: body: %v != %v", i, got, want)
		}
	}
}

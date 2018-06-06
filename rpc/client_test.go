package rpc

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"reflect"
	"testing"

	"github.com/ironzhang/zerone/rpc/codec"
	"github.com/ironzhang/zerone/rpc/codec/json_codec"
	"github.com/ironzhang/zerone/rpc/codes"
)

type testClientCodec struct {
	reqHeader codec.RequestHeader
	reqBody   interface{}

	respHeaderErr error
	respHeader    codec.ResponseHeader
	respBodyErr   error
	respBody      interface{}
}

func (c *testClientCodec) WriteRequest(h *codec.RequestHeader, a interface{}) error {
	c.reqHeader = *h
	c.reqBody = a
	return nil
}

func (c *testClientCodec) ReadResponseHeader(h *codec.ResponseHeader) error {
	if c.respHeaderErr != nil {
		return c.respHeaderErr
	}
	*h = c.respHeader
	return nil
}

func (c *testClientCodec) ReadResponseBody(a interface{}) error {
	if c.respBodyErr != nil {
		return c.respBodyErr
	}
	if a == nil {
		return nil
	}
	data, err := json.Marshal(c.respBody)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, a)
}

func (c *testClientCodec) Close() error {
	return nil
}

func TestClientReadResponseCorrect(t *testing.T) {
	tests := []struct {
		header codec.ResponseHeader
		body   interface{}
		reply  interface{}
	}{
		{
			header: codec.ResponseHeader{ClassMethod: "Arith.Add", Sequence: 1},
			body:   &Reply{3},
			reply:  &Reply{},
		},
		{
			header: codec.ResponseHeader{ClassMethod: "Arith.Div", Sequence: 1},
			body:   &Reply{3},
			reply:  &Reply{},
		},
	}

	for i, tt := range tests {
		clientCodec := &testClientCodec{respHeader: tt.header, respBody: tt.body}
		client := Client{codec: clientCodec}
		call := &Call{
			Reply: tt.reply,
			Done:  make(chan *Call, 1),
		}
		client.pending.Store(tt.header.Sequence, call)
		if _, err := client.readResponse(); err != nil {
			t.Fatalf("case%d: read response: %v", i, err)
		}
		<-call.Done
		if got, want := tt.reply, tt.body; !reflect.DeepEqual(got, want) {
			t.Fatalf("case%d: reply: %v != %v", i, got, want)
		}
		t.Logf("case%d: reply=%v", i, tt.reply)
	}
}

func TestClientReadResponseError(t *testing.T) {
	tests := []struct {
		headerErr error
		header    codec.ResponseHeader
		bodyErr   error
		expectErr error
	}{
		{
			headerErr: io.EOF,
			expectErr: io.EOF,
		},
		{
			header:    codec.ResponseHeader{ClassMethod: "Arith.Add", Sequence: 1},
			bodyErr:   io.ErrUnexpectedEOF,
			expectErr: io.ErrUnexpectedEOF,
		},
		{
			header: codec.ResponseHeader{
				ClassMethod: "Arith.Add",
				Sequence:    1,
				Error: codec.Error{
					Code:  int(codes.Internal),
					Desc:  codes.Internal.String(),
					Cause: "cause",
				},
			},
			expectErr: nil,
		},
	}

	for i, tt := range tests {
		clientCodec := &testClientCodec{respHeaderErr: tt.headerErr, respHeader: tt.header, respBodyErr: tt.bodyErr}
		client := Client{codec: clientCodec}
		call := &Call{Done: make(chan *Call, 1)}
		client.pending.Store(tt.header.Sequence, call)
		_, err := client.readResponse()
		if err != tt.expectErr {
			t.Fatalf("case%d: read response expect %v but got %v", i, tt.expectErr, err)
		}
		if err != nil {
			t.Logf("case%d: read response: %v", i, err)
		}
		select {
		case <-call.Done:
			t.Logf("case%d: error=%v", i, call.Error)
		default:
		}
	}
}

func CodecPipe() (codec.ClientCodec, codec.ServerCodec) {
	c, s := net.Pipe()
	return json_codec.NewClientCodec(c), json_codec.NewServerCodec(s)
}

func TestClientReading(t *testing.T) {
	tests := []struct {
		header codec.ResponseHeader
		body   interface{}
		reply  interface{}
	}{
		{
			header: codec.ResponseHeader{Sequence: 1},
			body:   &Reply{3},
			reply:  &Reply{},
		},
		{
			header: codec.ResponseHeader{Sequence: 1},
			body:   &Reply{3},
			reply:  &Reply{},
		},
	}

	clientCodec, serverCodec := CodecPipe()
	client := &Client{codec: clientCodec}
	call := &Call{Done: make(chan *Call, 1)}
	client.pending.Store(0, call)
	go func() {
		for i, tt := range tests {
			call := &Call{Reply: tt.reply, Done: make(chan *Call, 1)}
			client.pending.Store(tt.header.Sequence, call)
			serverCodec.WriteResponse(&tt.header, tt.body)
			<-call.Done
			if got, want := tt.reply, tt.body; !reflect.DeepEqual(got, want) {
				t.Errorf("case%d: reply: %v != %v", i, got, want)
			}
		}
		serverCodec.Close()
	}()
	client.reading()
	<-call.Done
	t.Logf("Error: %v", call.Error)
}

func TestClientCall(t *testing.T) {
	tests := []struct {
		classMethod string
		args        interface{}
		reply       interface{}
		result      interface{}
	}{
		{
			classMethod: "Arith.Add",
			args:        Args{1, 2},
			reply:       &Reply{},
			result:      &Reply{3},
		},
		{
			classMethod: "Arith.Div",
			args:        Args{10, 2},
			reply:       &Reply{},
			result:      &Reply{5},
		},
	}

	clientCodec, serverCodec := CodecPipe()
	go func() {
		var req codec.RequestHeader
		var resp codec.ResponseHeader
		for _, tt := range tests {
			serverCodec.ReadRequestHeader(&req)
			serverCodec.ReadRequestBody(nil)
			resp.ClassMethod = req.ClassMethod
			resp.Sequence = req.Sequence
			serverCodec.WriteResponse(&resp, tt.result)
		}
	}()
	client := NewClientWithCodec("client", clientCodec)
	defer client.Close()
	for _, tt := range tests {
		if err := client.Call(context.Background(), tt.classMethod, tt.args, tt.reply, 0); err != nil {
			t.Fatalf("call %q: %v", tt.classMethod, err)
		}
		if got, want := tt.reply, tt.result; !reflect.DeepEqual(got, want) {
			t.Fatalf("call %q reply: %v != %v", tt.classMethod, got, want)
		}
		t.Logf("call %q success, args=%v, reply=%v", tt.classMethod, tt.args, tt.reply)
	}
}

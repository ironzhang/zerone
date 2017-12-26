package rpc

import (
	"context"
	"encoding/json"
	"net"
	"reflect"
	"testing"

	"github.com/ironzhang/zerone/rpc/codec"
	"github.com/ironzhang/zerone/rpc/codec/json-codec"
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
	data, err := json.Marshal(c.respBody)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, a); err != nil {
		return err
	}
	return nil
}

func (c *testClientCodec) Close() error {
	return nil
}

func TestClientGo(t *testing.T) {
	tests := []struct {
		serviceMethod string
		args          interface{}
		reply         interface{}
		result        interface{}
	}{
		{
			serviceMethod: "Arith.Add",
			args:          Args{1, 2},
			reply:         &Reply{},
			result:        &Reply{3},
		},
	}

	cli, svr := net.Pipe()
	c := NewClient(cli)
	s := json_codec.NewServerCodec(svr)
	for _, tt := range tests {
		call, err := c.Go(context.Background(), tt.serviceMethod, tt.args, tt.reply, nil)
		if err != nil {
			t.Fatalf("call %q: %v", tt.serviceMethod, err)
		}
		h := codec.ResponseHeader{
			ServiceMethod: tt.serviceMethod,
			Sequence:      1,
		}
		s.WriteResponse(&h, tt.result)
		<-call.Done

		if got, want := tt.reply, tt.result; !reflect.DeepEqual(got, want) {
			t.Fatalf("reply: %v != %v", got, want)
		}
	}
}

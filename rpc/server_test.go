package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
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

type Arith int

// Some of Arith's methods have value args, some have pointer args. That's deliberate.

func (t *Arith) Add(ctx context.Context, args Args, reply *Reply) error {
	reply.C = args.A + args.B
	return nil
}

func (t *Arith) Mul(ctx context.Context, args *Args, reply *Reply) error {
	reply.C = args.A * args.B
	return nil
}

func (t *Arith) Div(ctx context.Context, args Args, reply *Reply) error {
	if args.B == 0 {
		return errors.New("divide by zero")
	}
	reply.C = args.A / args.B
	return nil
}

func (t *Arith) String(ctx context.Context, args *Args, reply *string) error {
	*reply = fmt.Sprintf("%d+%d=%d", args.A, args.B, args.A+args.B)
	return nil
}

func (t *Arith) Scan(ctx context.Context, args string, reply *Reply) (err error) {
	_, err = fmt.Sscan(args, &reply.C)
	return
}

func (t *Arith) Error(ctx context.Context, args *Args, reply *Reply) error {
	panic("ERROR")
}

type BuiltinTypes struct{}

func (BuiltinTypes) Map(ctx context.Context, args *Args, reply *map[int]int) error {
	(*reply)[args.A] = args.B
	return nil
}

func (BuiltinTypes) Slice(ctx context.Context, args *Args, reply *[]int) error {
	*reply = append(*reply, args.A, args.B)
	return nil
}

func (BuiltinTypes) Array(ctx context.Context, args *Args, reply *[2]int) error {
	(*reply)[0] = args.A
	(*reply)[1] = args.B
	return nil
}

type hidden int

func (t *hidden) Exported(ctx context.Context, args Args, reply *Reply) error {
	reply.C = args.A + args.B
	return nil
}

type Embed struct {
	hidden
}

func TestServerRegisterCorrect(t *testing.T) {
	var arith Arith
	var embed Embed
	var builtinTypes BuiltinTypes
	tests := []struct {
		rcvr    interface{}
		name    string
		service string
	}{
		{rcvr: &arith, name: "arith", service: "arith"},
		{rcvr: &embed, name: "", service: "Embed"},
		{rcvr: builtinTypes, name: "Builtin", service: "Builtin"},
	}

	var s Server
	for i, tt := range tests {
		if err := s.register(tt.rcvr, tt.name); err != nil {
			t.Fatalf("case%d: register: %v", i, err)
		}
		if _, ok := s.serviceMap.Load(tt.service); !ok {
			t.Fatalf("case%d: %q not found", i, tt.service)
		}
	}
}

func TestServerRegisterError(t *testing.T) {
	var s Server

	// 未导出
	var h hidden
	if err := s.register(&h, ""); err == nil {
		t.Errorf("register return error is nil")
	} else {
		t.Logf("register: %v", err)
	}

	// 没有方法
	type A struct{}
	if err := s.register(A{}, ""); err == nil {
		t.Errorf("register return error is nil")
	} else {
		t.Logf("register: %v", err)
	}

	// 重复注册
	var a Arith
	if err := s.register(&a, ""); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := s.register(&a, ""); err == nil {
		t.Errorf("register return error is nil")
	} else {
		t.Logf("register: %v", err)
	}
}

func TestSplitServiceMethodCorrect(t *testing.T) {
	tests := []struct {
		serviceMethod string
		service       string
		method        string
	}{
		{serviceMethod: ".", service: "", method: ""},
		{serviceMethod: "a.", service: "a", method: ""},
		{serviceMethod: ".a", service: "", method: "a"},
		{serviceMethod: "a.b", service: "a", method: "b"},
		{serviceMethod: "ABC.abc", service: "ABC", method: "abc"},
	}
	for _, tt := range tests {
		service, method, err := splitServiceMethod(tt.serviceMethod)
		if err != nil {
			t.Fatalf("serviceMethod=%q: splitServiceMethod: %v", tt.serviceMethod, err)
		}
		if got, want := service, tt.service; got != want {
			t.Errorf("serviceMethod=%q: service: %q != %q", tt.serviceMethod, got, want)
		}
		if got, want := method, tt.method; got != want {
			t.Errorf("serviceMethod=%q: method: %q != %q", tt.serviceMethod, got, want)
		}
	}
}

func TestSplitServiceMethodError(t *testing.T) {
	tests := []string{"", "a", "aaa", "a@b"}
	for _, tt := range tests {
		_, _, err := splitServiceMethod(tt)
		if err == nil {
			t.Fatalf("serviceMethod=%q: splitServiceMethod return error is nil", tt)
		} else {
			t.Logf("serviceMethod=%q: splitServiceMethod: %v", tt, err)
		}
	}
}

func getMethod(rcvr interface{}, methodName string) *method {
	m, ok := reflect.TypeOf(rcvr).MethodByName(methodName)
	if !ok {
		panic(fmt.Errorf("%s method not found", methodName))
	}
	return &method{
		method: m,
		args:   m.Type.In(2),
		reply:  m.Type.In(3),
	}
}

func TestLookupServiceMethod(t *testing.T) {
	var a Arith
	var b BuiltinTypes
	var s Server
	if err := s.Register(&a); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if err := s.Register(b); err != nil {
		t.Fatalf("Register: %v", err)
	}

	tests := []struct {
		service string
		method  string
		rcvr    reflect.Value
		meth    *method
	}{
		{service: "Arith", method: "Add", rcvr: reflect.ValueOf(&a), meth: getMethod(&a, "Add")},
		{service: "Arith", method: "Div", rcvr: reflect.ValueOf(&a), meth: getMethod(&a, "Div")},
		{service: "BuiltinTypes", method: "Map", rcvr: reflect.ValueOf(b), meth: getMethod(b, "Map")},
		{service: "BuiltinTypes", method: "Slice", rcvr: reflect.ValueOf(b), meth: getMethod(b, "Slice")},
	}
	for _, tt := range tests {
		rcvr, meth, err := s.lookupServiceMethod(tt.service, tt.method)
		if err != nil {
			t.Fatalf("service=%q, method=%q: lookupServiceMethod: %v", tt.service, tt.method, err)
		}
		if got, want := rcvr, tt.rcvr; got != want {
			t.Fatalf("service=%q, method=%q: rcvr: %v != %v", tt.service, tt.method, got, want)
		}
		if got, want := meth.method.Name, tt.meth.method.Name; got != want {
			t.Fatalf("service=%q, method=%q: method: %v != %v", tt.service, tt.method, got, want)
		}
		if got, want := meth.args, tt.meth.args; got != want {
			t.Fatalf("service=%q, method=%q: args: %v != %v", tt.service, tt.method, got, want)
		}
		if got, want := meth.reply, tt.meth.reply; got != want {
			t.Fatalf("service=%q, method=%q: reply: %v != %v", tt.service, tt.method, got, want)
		}
	}
}

type testServerCodec struct {
	reqHeaderErr error
	reqHeader    codec.RequestHeader
	reqBodyErr   error
	reqBody      interface{}

	respHeader codec.ResponseHeader
	respBody   interface{}
}

func (c *testServerCodec) ReadRequestHeader(h *codec.RequestHeader) error {
	if c.reqHeaderErr != nil {
		return c.reqHeaderErr
	}
	*h = c.reqHeader
	return nil
}

func (c *testServerCodec) ReadRequestBody(a interface{}) error {
	if c.reqBodyErr != nil {
		return c.reqBodyErr
	}
	data, err := json.Marshal(c.reqBody)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, a); err != nil {
		return err
	}
	return nil
}

func (c *testServerCodec) WriteResponse(h *codec.ResponseHeader, a interface{}) error {
	c.respHeader = *h
	c.respBody = a
	return nil
}

func TestServerReadRequestCorrect(t *testing.T) {
	var a Arith
	var b BuiltinTypes
	var e Embed
	var s Server
	if err := s.Register(&a); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if err := s.Register(b); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if err := s.Register(&e); err != nil {
		t.Fatalf("Register: %v", err)
	}

	tests := []struct {
		service string
		method  string
		rcvr    interface{}
		args    interface{}
		reply   interface{}
	}{
		{service: "Arith", method: "Add", rcvr: &a, args: Args{1, 2}, reply: &Reply{}},
		{service: "Arith", method: "Mul", rcvr: &a, args: &Args{1, 2}, reply: &Reply{}},
		{service: "BuiltinTypes", method: "Map", rcvr: b, args: &Args{1, 2}, reply: &map[int]int{}},
		{service: "BuiltinTypes", method: "Slice", rcvr: b, args: &Args{1, 2}, reply: &[]int{}},
		{service: "BuiltinTypes", method: "Array", rcvr: b, args: &Args{1, 2}, reply: &[2]int{}},
		{service: "Embed", method: "Exported", rcvr: &e, args: Args{1, 2}, reply: &Reply{}},
	}
	for _, tt := range tests {
		header := codec.RequestHeader{
			ServiceMethod: fmt.Sprintf("%s.%s", tt.service, tt.method),
			Sequence:      rand.Uint64(),
			TraceID:       "TraceID",
			ClientName:    "ClientName",
			Verbose:       true,
		}
		codec := &testServerCodec{reqHeader: header, reqBody: tt.args}

		req, method, rcvr, args, reply, err := s.readRequest(codec)
		if err != nil {
			t.Fatalf("readRequest: %v", err)
		}
		if got, want := *req, header; got != want {
			t.Fatalf("%s.%s: header: %v != %v", tt.service, tt.method, got, want)
		}
		if got, want := method.Name, tt.method; got != want {
			t.Fatalf("%s.%s: method: %v != %v", tt.service, tt.method, got, want)
		}
		if got, want := rcvr.Interface(), tt.rcvr; got != want {
			t.Fatalf("%s.%s: rcvr: %v != %v", tt.service, tt.method, got, want)
		}
		if got, want := args.Interface(), tt.args; !reflect.DeepEqual(got, want) {
			t.Fatalf("%s.%s: args: %v != %v", tt.service, tt.method, got, want)
		}
		if got, want := reply.Interface(), tt.reply; !reflect.DeepEqual(got, want) {
			t.Fatalf("%s.%s: reply: %#v != %#v", tt.service, tt.method, got, want)
		}
		t.Logf("req=%v, method=%s, rcvr=%v, args=%v, reply=%v", *req, method.Name, rcvr.Interface(), args.Interface(), reply.Interface())
	}
}

func TestServerReadRequestError(t *testing.T) {
	var a Arith
	var s Server
	if err := s.Register(&a); err != nil {
		t.Fatalf("Register: %v", err)
	}

	tests := []struct {
		headerErr error
		header    codec.RequestHeader
		bodyErr   error
		body      interface{}
		expectReq *codec.RequestHeader
	}{
		{
			headerErr: io.EOF,
			header:    codec.RequestHeader{},
			bodyErr:   nil,
			body:      nil,
			expectReq: nil,
		},
		{
			headerErr: nil,
			header: codec.RequestHeader{
				ServiceMethod: "",
				Sequence:      1,
			},
			bodyErr: nil,
			body:    nil,
			expectReq: &codec.RequestHeader{
				ServiceMethod: "",
				Sequence:      1,
			},
		},
		{
			headerErr: nil,
			header: codec.RequestHeader{
				ServiceMethod: ".Add",
				Sequence:      2,
			},
			bodyErr: nil,
			body:    nil,
			expectReq: &codec.RequestHeader{
				ServiceMethod: ".Add",
				Sequence:      2,
			},
		},
		{
			headerErr: nil,
			header: codec.RequestHeader{
				ServiceMethod: "Arith.",
				Sequence:      3,
			},
			bodyErr: nil,
			body:    nil,
			expectReq: &codec.RequestHeader{
				ServiceMethod: "Arith.",
				Sequence:      3,
			},
		},
		{
			headerErr: nil,
			header: codec.RequestHeader{
				ServiceMethod: "Arith.Add",
				Sequence:      4,
			},
			bodyErr: io.EOF,
			body:    nil,
			expectReq: &codec.RequestHeader{
				ServiceMethod: "Arith.Add",
				Sequence:      4,
			},
		},
	}
	for _, tt := range tests {
		codec := &testServerCodec{reqHeaderErr: tt.headerErr, reqHeader: tt.header, reqBodyErr: tt.bodyErr, reqBody: tt.body}

		req, _, _, _, _, err := s.readRequest(codec)
		if err == nil {
			t.Fatalf("readRequest: return error is nil")
		} else {
			t.Logf("readRequest: %v", err)
		}
		if got, want := req, tt.expectReq; !reflect.DeepEqual(got, want) {
			t.Fatalf("header: %+v != %+v", got, want)
		}
	}
}

package rpc

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
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

type hidden int

func (t *hidden) Exported(ctx context.Context, args Args, reply *Reply) error {
	reply.C = args.A + args.B
	return nil
}

type Embed struct {
	hidden
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
package rpc_test

import (
	"context"
	"errors"
	"net"
	"reflect"
	"testing"

	"github.com/ironzhang/zerone/rpc"
)

type Args struct {
	A, B int
}

type Quotient struct {
	Quo, Rem int
}

type Arith int

func (t *Arith) Multiply(ctx context.Context, args *Args, reply *int) error {
	*reply = args.A * args.B
	return nil
}

func (t *Arith) Divide(ctx context.Context, args *Args, quo *Quotient) error {
	if args.B == 0 {
		return errors.New("divide by zero")
	}
	quo.Quo = args.A / args.B
	quo.Rem = args.A % args.B
	return nil
}

func ServeRPC(network, address string) {
	ln, err := net.Listen(network, address)
	if err != nil {
		panic(err)
	}

	svr := rpc.NewServer("TestServer")
	if err = svr.Register(new(Arith)); err != nil {
		panic(err)
	}

	go svr.Accept(ln)
}

func init() {
	ServeRPC("tcp", ":2000")
}

func TestCallCorrect(t *testing.T) {
	c, err := rpc.Dial("tcp", "localhost:2000")
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Close()

	var reply int
	var result int = 6
	tests := []struct {
		serviceMethod string
		args          interface{}
		reply         interface{}
		result        interface{}
	}{
		{
			serviceMethod: "Arith.Multiply",
			args:          Args{A: 2, B: 3},
			reply:         nil,
			result:        nil,
		},
		{
			serviceMethod: "Arith.Multiply",
			args:          Args{A: 2, B: 3},
			reply:         &reply,
			result:        &result,
		},
		{
			serviceMethod: "Arith.Divide",
			args:          Args{A: 2, B: 3},
			reply:         nil,
			result:        nil,
		},
		{
			serviceMethod: "Arith.Divide",
			args:          Args{A: 2, B: 3},
			reply:         &Quotient{},
			result:        &Quotient{Quo: 2 / 3, Rem: 2 % 3},
		},
	}
	for i, tt := range tests {
		if err := c.Call(context.Background(), tt.serviceMethod, tt.args, tt.reply); err != nil {
			t.Fatalf("case%d: call: %v", i, err)
		}
		if got, want := tt.reply, tt.result; !reflect.DeepEqual(got, want) {
			t.Fatalf("case%d: reply: %v != %v", i, got, want)
		}
	}
}

func TestCallError(t *testing.T) {
	c, err := rpc.Dial("tcp", "localhost:2000")
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Close()

	tests := []struct {
		serviceMethod string
		args          interface{}
		reply         interface{}
	}{
		{
			serviceMethod: "Arith.Divide",
			args:          Args{A: 2, B: 0},
			reply:         &Quotient{},
		},
		{
			serviceMethod: "Arith.Divide",
			args:          Args{A: 2, B: 0},
			reply:         &Quotient{},
		},
	}
	for i, tt := range tests {
		if err := c.Call(context.Background(), tt.serviceMethod, tt.args, tt.reply); err != nil {
			t.Logf("case%d: call: %v", i, err)
		}
	}
}

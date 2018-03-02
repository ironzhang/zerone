package rpc_test

import (
	"context"
	"errors"
	"io/ioutil"
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
	svr.SetTraceOutput(ioutil.Discard)

	go svr.Accept(ln)
}

func init() {
	ServeRPC("tcp", ":2000")
}

func TestCallCorrect(t *testing.T) {
	c, err := rpc.Dial("TestCallCorrect", "tcp", "localhost:2000")
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
	c, err := rpc.Dial("TestCallError", "tcp", "localhost:2000")
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	c.SetTraceOutput(ioutil.Discard)
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

func TestGo(t *testing.T) {
	c, err := rpc.Dial("TestGo", "tcp", "localhost:2000")
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	c.SetTraceOutput(ioutil.Discard)
	defer c.Close()

	var reply int
	tests := []struct {
		serviceMethod string
		args          interface{}
		reply         interface{}
	}{
		{
			serviceMethod: "Arith.Multiply",
			args:          Args{A: 2, B: 3},
			reply:         nil,
		},
		{
			serviceMethod: "Arith.Multiply",
			args:          Args{A: 2, B: 3},
			reply:         &reply,
		},
		{
			serviceMethod: "Arith.Divide",
			args:          Args{A: 2, B: 3},
			reply:         nil,
		},
		{
			serviceMethod: "Arith.Divide",
			args:          Args{A: 2, B: 3},
			reply:         &Quotient{},
		},
		{
			serviceMethod: "Arith.Divide",
			args:          Args{A: 2, B: 0},
			reply:         nil,
		},
	}
	done := make(chan *rpc.Call, len(tests))
	for i, tt := range tests {
		_, err := c.Go(context.Background(), tt.serviceMethod, tt.args, tt.reply, done)
		if err != nil {
			t.Fatalf("case%d: call: %v", i, err)
		}
	}
	for range tests {
		call := <-done
		t.Logf("%q: error=%v, reply=%v", call.Header.ServiceMethod, call.Error, call.Reply)
	}
}

func BenchmarkOneClientSerialCall(b *testing.B) {
	c, err := rpc.Dial("BenchmarkOneClientSerialCall", "tcp", "localhost:2000")
	if err != nil {
		b.Fatalf("dial: %v", err)
	}
	defer c.Close()

	var reply int
	var args = Args{4, 4}
	var result = args.A * args.B
	var ctx = context.Background()
	for i := 0; i < b.N; i++ {
		if err = c.Call(ctx, "Arith.Multiply", &args, &reply); err != nil {
			b.Errorf("Call Arith.Multiply: %v", err)
		}
		if reply != result {
			b.Errorf("Result: %d != %d * %d", reply, args.A, args.B)
		}
	}
}

func BenchmarkOneClientParallelCall(b *testing.B) {
	c, err := rpc.Dial("BenchmarkOneClientParallelCall", "tcp", "localhost:2000")
	if err != nil {
		b.Fatalf("dial: %v", err)
	}
	defer c.Close()

	b.SetParallelism(10)
	b.RunParallel(func(pb *testing.PB) {
		var err error
		var reply int
		var args = Args{4, 4}
		var result = args.A * args.B
		var ctx = context.Background()
		for pb.Next() {
			if err = c.Call(ctx, "Arith.Multiply", &args, &reply); err != nil {
				b.Errorf("Call Arith.Multiply: %v", err)
			}
			if reply != result {
				b.Errorf("Result: %d != %d * %d", reply, args.A, args.B)
			}
		}
	})
}

func BenchmarkNClientsCall(b *testing.B) {
	b.SetParallelism(100)
	b.RunParallel(func(pb *testing.PB) {
		c, err := rpc.Dial("BenchmarkNClientsCall", "tcp", "localhost:2000")
		if err != nil {
			b.Fatalf("dial: %v", err)
		}
		defer c.Close()

		var reply int
		var args = Args{4, 4}
		var result = args.A * args.B
		var ctx = context.Background()
		for pb.Next() {
			if err = c.Call(ctx, "Arith.Multiply", &args, &reply); err != nil {
				b.Errorf("Call Arith.Multiply: %v", err)
			}
			if reply != result {
				b.Errorf("Result: %d != %d * %d", reply, args.A, args.B)
			}
		}
	})
}

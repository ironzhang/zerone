package rpc_test

import (
	"context"
	"errors"
	"io/ioutil"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/ironzhang/zerone/rpc"
	"github.com/ironzhang/zerone/rpc/trace"
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

func (t *Arith) Random(ctx context.Context, args interface{}, reply *int) error {
	*reply = 6
	return nil
}

type Time int

func (t *Time) Sleep(ctx context.Context, args int, reply interface{}) error {
	d := time.Duration(args) * time.Millisecond
	time.Sleep(d)
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
	if err = svr.Register(new(Time)); err != nil {
		panic(err)
	}
	svr.SetTraceOutput(trace.NewStdOutput(ioutil.Discard))

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
	c.SetTraceOutput(trace.NewStdOutput(ioutil.Discard))
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
		{
			serviceMethod: "Arith.Random",
			args:          nil,
			reply:         &reply,
			result:        &result,
		},
	}
	for i, tt := range tests {
		if err := c.Call(context.Background(), tt.serviceMethod, tt.args, tt.reply, 0); err != nil {
			t.Fatalf("case%d: call: %v", i, err)
		}
		if got, want := tt.reply, tt.result; !reflect.DeepEqual(got, want) {
			t.Fatalf("case%d: reply: %#v != %#v", i, got, want)
		}
	}
}

func TestCallError(t *testing.T) {
	c, err := rpc.Dial("TestCallError", "tcp", "localhost:2000")
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	c.SetTraceOutput(trace.NewStdOutput(ioutil.Discard))
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
		if err := c.Call(context.Background(), tt.serviceMethod, tt.args, tt.reply, 0); err != nil {
			t.Logf("case%d: call: %v", i, err)
		}
	}
}

func TestGo(t *testing.T) {
	c, err := rpc.Dial("TestGo", "tcp", "localhost:2000")
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	c.SetTraceOutput(trace.NewStdOutput(ioutil.Discard))
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
		_, err := c.Go(context.Background(), tt.serviceMethod, tt.args, tt.reply, 0, done)
		if err != nil {
			t.Fatalf("case%d: go: %v", i, err)
		}
	}
	for range tests {
		call := <-done
		t.Logf("%q: error=%v, reply=%v", call.Header.ClassMethod, call.Error, call.Reply)
	}
}

func TestGoTimeout(t *testing.T) {
	c, err := rpc.Dial("TestGo", "tcp", "localhost:2000")
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	c.SetTraceOutput(trace.NewStdOutput(ioutil.Discard))
	defer c.Close()

	tests := []struct {
		sleep   int
		timeout time.Duration
		err     error
	}{
		{sleep: 100, timeout: time.Second, err: nil},
		{sleep: 100, timeout: 50 * time.Millisecond, err: rpc.ErrTimeout},
	}
	for i, tt := range tests {
		call, err := c.Go(context.Background(), "Time.Sleep", tt.sleep, nil, tt.timeout, nil)
		if err != nil {
			t.Fatalf("%d: go: %v", i, err)
		}
		<-call.Done
		if got, want := call.Error, tt.err; got != want {
			t.Fatalf("%d: error: got %v, want %v", i, got, want)
		} else {
			t.Logf("%d: error: got %v", i, got)
		}
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
		if err = c.Call(ctx, "Arith.Multiply", &args, &reply, 0); err != nil {
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
			if err = c.Call(ctx, "Arith.Multiply", &args, &reply, 0); err != nil {
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
			if err = c.Call(ctx, "Arith.Multiply", &args, &reply, 0); err != nil {
				b.Errorf("Call Arith.Multiply: %v", err)
			}
			if reply != result {
				b.Errorf("Result: %d != %d * %d", reply, args.A, args.B)
			}
		}
	})
}

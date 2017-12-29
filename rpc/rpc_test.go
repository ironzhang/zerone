package rpc_test

import (
	"context"
	"errors"
	"net"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
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

func TestGo(t *testing.T) {
	c, err := rpc.Dial("tcp", "localhost:2000")
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
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
		t.Logf("%q: error=%v, reply=%v", call.ServiceMethod, call.Error, call.Reply)
	}
}

func benchmarkClientCall(b *testing.B, c *rpc.Client) {
	b.StopTimer()
	var N = int64(b.N)
	var args = Args{4, 2}
	var procs = runtime.GOMAXPROCS(-1) * 10
	b.StartTimer()

	var wg sync.WaitGroup
	wg.Add(procs)
	for i := 0; i < procs; i++ {
		go func() {
			var err error
			var reply int
			for atomic.AddInt64(&N, -1) >= 0 {
				if err = c.Call(context.Background(), "Arith.Multiply", args, &reply); err != nil {
					b.Errorf("client call: %v", err)
				}
				if reply != args.A*args.B {
					b.Errorf("reply: %d != %d * %d", reply, args.A, args.B)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func BenchmarkOneClientCall(b *testing.B) {
	c, err := rpc.Dial("tcp", "localhost:2000")
	if err != nil {
		b.Fatalf("dial: %v", err)
	}
	defer c.Close()
	benchmarkClientCall(b, c)
}

func DialN(network, address string, n int) ([]*rpc.Client, func(), error) {
	var err error
	var clients = make([]*rpc.Client, n)
	for i := 0; i < n; i++ {
		clients[i], err = rpc.Dial(network, address)
		if err != nil {
			return nil, nil, err
		}
	}

	destroy := func() {
		for _, c := range clients {
			c.Close()
		}
	}
	return clients, destroy, nil
}

func BenchmarkMultClientCall(b *testing.B) {
	b.StopTimer()
	n := 1000
	clients, destroy, err := DialN("tcp", "localhost:2000", n)
	if err != nil {
		b.Fatalf("dial n clients: %v", err)
	}
	b.StartTimer()

	var args = Args{4, 2}
	var reply int
	var wg sync.WaitGroup
	wg.Add(b.N)
	for i := 0; i < b.N; i++ {
		go func(c *rpc.Client) {
			defer wg.Done()
			if err = c.Call(context.Background(), "Arith.Multiply", args, &reply); err != nil {
				b.Fatalf("call: %v", err)
			}
		}(clients[i%n])
	}
	wg.Wait()

	b.StopTimer()
	destroy()
	b.StartTimer()
}

package zserver

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ironzhang/x-pearls/govern"
	"github.com/ironzhang/x-pearls/govern/stub"
	"github.com/ironzhang/zerone/pkg/endpoint"
	"github.com/ironzhang/zerone/rpc"
)

type Echo struct{}

func (Echo) Echo(ctx context.Context, req string, reply *string) error {
	*reply = req
	return nil
}

func OpenTestDriver() govern.Driver {
	d, err := govern.Open(stub.DriverName, "test", nil)
	if err != nil {
		panic(err)
	}
	return d
}

func WaitEndpoints(c govern.Consumer, n int) error {
	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		if len(c.GetEndpoints()) == n {
			return nil
		}
	}
	return errors.New("timeout")
}

func TestServerWithDriver(t *testing.T) {
	d := OpenTestDriver()
	c := d.NewConsumer("TestServer", &endpoint.Endpoint{}, nil)
	s := New("TestServer-0", "TestServer", d)
	if err := s.Register(Echo{}); err != nil {
		t.Fatalf("register: %v", err)
	}

	tests := []struct {
		net  string
		addr string
	}{
		{net: "tcp", addr: "localhost:5000"},
		{net: "tcp", addr: "localhost:5001"},
	}
	for _, tt := range tests {
		go func(net, addr string) {
			if err := s.ListenAndServe(net, addr); err != nil {
				t.Fatalf("listen and serve: %v", err)
			}
		}(tt.net, tt.addr)
	}

	if err := WaitEndpoints(c, len(tests)); err != nil {
		t.Fatalf("wait endpoints: %v", err)
	}

	for i, p := range c.GetEndpoints() {
		ep := p.(*endpoint.Endpoint)
		rc, err := rpc.Dial(fmt.Sprintf("TestClient-%d", i), ep.Net, ep.Addr)
		if err != nil {
			t.Fatalf("%d: dial: %v", i, err)
		}
		defer rc.Close()
		//rc.SetTraceVerbose(1)

		req, reply := "hello", ""
		if err = rc.Call(context.Background(), "Echo.Echo", req, &reply, 0); err != nil {
			t.Fatalf("%d: call: %v", i, err)
		}
		if got, want := reply, req; got != want {
			t.Fatalf("%d: reply: got %v, want %v", i, got, want)
		} else {
			t.Logf("%d: reply: got %v", i, got)
		}
	}

	s.Close()
	if err := WaitEndpoints(c, 0); err != nil {
		t.Fatalf("wait endpoints: %v", err)
	} else {
		t.Logf("wait endpoints: n=0")
	}
}

func TestServerWithoutDriver(t *testing.T) {
	s := New("TestServer-0", "TestServer", nil)
	if err := s.Register(Echo{}); err != nil {
		t.Fatalf("register: %v", err)
	}

	tests := []struct {
		net  string
		addr string
	}{
		{net: "tcp", addr: "localhost:6000"},
		{net: "tcp", addr: "localhost:6001"},
	}
	for _, tt := range tests {
		go func(net, addr string) {
			if err := s.ListenAndServe(net, addr); err != nil {
				t.Fatalf("listen and serve: %v", err)
			}
		}(tt.net, tt.addr)
	}
	time.Sleep(500 * time.Millisecond)

	for i, tt := range tests {
		rc, err := rpc.Dial(fmt.Sprintf("TestClient-%d", i), tt.net, tt.addr)
		if err != nil {
			t.Fatalf("%d: dial: %v", i, err)
		}
		defer rc.Close()
		//rc.SetTraceVerbose(1)

		req, reply := "hello", ""
		if err = rc.Call(context.Background(), "Echo.Echo", req, &reply, 0); err != nil {
			t.Fatalf("%d: call: %v", i, err)
		}
		if got, want := reply, req; got != want {
			t.Fatalf("%d: reply: got %v, want %v", i, got, want)
		} else {
			t.Logf("%d: reply: got %v", i, got)
		}
	}

	s.Close()
}

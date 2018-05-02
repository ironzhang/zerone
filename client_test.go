package zerone

import (
	"context"
	"net"
	"testing"

	"github.com/ironzhang/zerone/route"
	"github.com/ironzhang/zerone/route/tables/stable"
	"github.com/ironzhang/zerone/rpc"
)

type Echo int

func (p *Echo) Echo(ctx context.Context, args string, reply *string) error {
	*reply = args
	return nil
}

func (p *Echo) Inc(ctx context.Context, args interface{}, reply *int) error {
	*reply = int(*p)
	(*p)++
	return nil
}

func ServeEcho(network, address string) {
	ln, err := net.Listen(network, address)
	if err != nil {
		panic(err)
	}

	svr := rpc.NewServer("Server")
	if err = svr.Register(new(Echo)); err != nil {
		panic(err)
	}

	go svr.Accept(ln)
}

func init() {
	ServeEcho("tcp", ":4000")
}

func TestClientCall(t *testing.T) {
	tb := stable.NewTable([]route.Endpoint{
		{"0", "tcp", "localhost:4000", 0},
	})
	c := NewClient("Client", tb)
	defer c.Close()

	args, reply := "hello, world", ""
	err := c.Call(context.Background(), nil, "Echo.Echo", args, &reply, 0)
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if args != reply {
		t.Errorf("args(%s) != reply(%s)", args, reply)
	} else {
		t.Logf("args(%s) == reply(%s)", args, reply)
	}
}

func TestClientBroadcast(t *testing.T) {
	tb := stable.NewTable([]route.Endpoint{
		{"0", "tcp", "localhost:4000", 0},
		{"1", "tcp", "localhost:4000", 0},
		{"2", "tcp", "localhost:4000", 0},
	})
	c := NewClient("Client", tb)
	defer c.Close()

	var reply int
	ch := c.Broadcast(context.Background(), "Echo.Inc", nil, reply, 0)
	for res := range ch {
		if res.Error != nil {
			t.Errorf("call %v error: %v", res.Endpoint, res.Error)
			continue
		}
		t.Logf("call %v reply: %v", res.Endpoint, *res.Reply.(*int))
	}
}

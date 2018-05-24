package main

import (
	"flag"
	"net"

	"github.com/ironzhang/x-pearls/zlog"
	"github.com/ironzhang/zerone/examples/rpc/arith"
	"github.com/ironzhang/zerone/examples/rpc/trace"
	"github.com/ironzhang/zerone/rpc"
)

type Options struct {
	net  string
	addr string
}

func (o *Options) Parse() {
	flag.StringVar(&o.net, "net", "tcp", "network")
	flag.StringVar(&o.addr, "addr", ":10000", "address")
	flag.Parse()
}

func main() {
	var opts Options
	opts.Parse()

	ln, err := net.Listen(opts.net, opts.addr)
	if err != nil {
		zlog.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	s := rpc.NewServer("ArithServer")
	if err := s.Register(arith.Arith(0)); err != nil {
		zlog.Fatalf("register: %v", err)
	}
	if err := s.Register(trace.New(s)); err != nil {
		zlog.Fatalf("register: %v", err)
	}
	zlog.Infof("serve on %s@%s", opts.net, opts.addr)
	s.Accept(ln)
}

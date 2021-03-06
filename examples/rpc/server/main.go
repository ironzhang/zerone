package main

import (
	"flag"
	"net"

	log "github.com/ironzhang/tlog"
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
		log.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	s := rpc.NewServer("ArithServer")
	if err := s.Register(arith.Arith(0)); err != nil {
		log.Fatalf("register: %v", err)
	}
	if err := s.Register(trace.New(s)); err != nil {
		log.Fatalf("register: %v", err)
	}
	log.Infof("serve on %s@%s", opts.net, opts.addr)
	s.Accept(ln)
}

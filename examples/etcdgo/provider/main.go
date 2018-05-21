package main

import (
	"flag"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/ironzhang/zerone/route"
	"github.com/ironzhang/zerone/route/governance/etcdgo"
	"github.com/ironzhang/zerone/zlog"
)

type Options struct {
	Level    int
	Node     string
	RandLoad bool
}

func (o *Options) Parse() {
	flag.IntVar(&o.Level, "level", int(zlog.DEBUG), "log level")
	flag.StringVar(&o.Node, "node", "node1", "node name")
	flag.BoolVar(&o.RandLoad, "rand-load", false, "rand load")
	flag.Parse()
}

func (o *Options) Endpoint() route.Endpoint {
	ep := route.Endpoint{
		Name: o.Node,
		Net:  "tcp",
		Addr: "localhost:2000",
	}
	if o.RandLoad {
		ep.Load = rand.Float64()
	}
	return ep
}

func main() {
	var opts Options
	opts.Parse()
	zlog.Default.SetLevel(zlog.Level(opts.Level))

	c, err := etcdgo.NewClient("test", client.Config{Endpoints: []string{"http://127.0.0.1:2379"}})
	if err != nil {
		zlog.Fatalw("new client", "error", err)
	}
	defer time.Sleep(time.Second)

	p := c.NewProvider("ac-test", opts.Endpoint)
	defer p.Close()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
}

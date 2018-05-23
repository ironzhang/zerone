package main

import (
	"flag"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/ironzhang/zerone/govern"
	"github.com/ironzhang/zerone/govern/etcdv2"
	"github.com/ironzhang/zerone/route"
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

func main() {
	var opts Options
	opts.Parse()
	zlog.Default.SetLevel(zlog.Level(opts.Level))

	d, err := govern.Open(etcdv2.DriverName, "test", client.Config{Endpoints: []string{"http://127.0.0.1:2379"}})
	if err != nil {
		zlog.Fatalw("open", "error", err)
	}
	defer d.Close()
	defer time.Sleep(time.Second)

	p := d.NewProvider("ac-test", 5*time.Second, func() govern.Endpoint {
		ep := &route.Endpoint{
			Name: opts.Node,
			Net:  "tcp",
			Addr: "localhost:2000",
		}
		if opts.RandLoad {
			ep.Load = rand.Float64()
		}
		return ep
	})
	defer p.Close()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
}

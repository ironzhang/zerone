package main

import (
	"flag"
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
	Level int
}

func (o *Options) Parse() {
	flag.IntVar(&o.Level, "level", int(zlog.INFO), "log level")
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

	p := d.NewConsumer("ac-test", &route.Endpoint{}, refresh)
	defer p.Close()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
}

func refresh(endpoints []govern.Endpoint) {
	zlog.Infow("refresh", "endpoints", endpoints)
}

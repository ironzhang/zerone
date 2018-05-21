package main

import (
	"flag"
	"os"
	"os/signal"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/ironzhang/zerone/route"
	"github.com/ironzhang/zerone/route/tables/dtable/etcd"
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

	c, err := etcd.NewClient("test", client.Config{Endpoints: []string{"http://127.0.0.1:2379"}})
	if err != nil {
		zlog.Fatalw("new client", "error", err)
	}
	defer time.Sleep(time.Second)

	p := c.NewConsumer("ac-test", refresh)
	defer p.Close()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
}

func refresh(endpoints []route.Endpoint) {
	zlog.Infow("refresh", "endpoints", endpoints)
}

package main

import (
	"os"
	"os/signal"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/ironzhang/x-pearls/govern/etcdv2"
	"github.com/ironzhang/x-pearls/zlog"
	"github.com/ironzhang/zerone"
	"github.com/ironzhang/zerone/examples/rpc/arith"
)

func main() {
	defer time.Sleep(10 * time.Millisecond)

	opts := zerone.Options{
		Zerone: "DZerone",
		DOptions: zerone.DOptions{
			Namespace: "zerone",
			Driver:    etcdv2.DriverName,
			Config:    client.Config{Endpoints: []string{"http://127.0.0.1:2379"}},
		},
	}
	z, err := zerone.NewZerone(opts)
	if err != nil {
		zlog.Fatalf("new zerone: %v", err)
	}
	defer z.Close()

	svr, err := z.NewServer("S1", "Arith")
	if err != nil {
		zlog.Fatalf("new server: %v", err)
	}
	defer svr.Close()

	if err = svr.Register(arith.Arith(0)); err != nil {
		zlog.Fatalf("register: %v", err)
	}

	go func() {
		net, addr := "tcp", "localhost:8000"
		zlog.Infof("listen and serve on %s://%s", net, addr)
		if err = svr.ListenAndServe(net, addr); err != nil {
			zlog.Fatalf("listen and serve: %v", err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
}

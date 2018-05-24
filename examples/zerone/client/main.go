package main

import (
	"context"
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

	c, err := z.NewClient("C1", "Arith")
	if err != nil {
		zlog.Fatalf("new client: %v", err)
	}
	defer c.Close()

	time.Sleep(10 * time.Millisecond)

	var reply int
	if err = c.Call(context.Background(), nil, "Arith.Add", arith.Args{1, 2}, &reply, 0); err != nil {
		zlog.Fatalf("call: %v", err)
	}
	zlog.Infof("reply: %v", reply)
}

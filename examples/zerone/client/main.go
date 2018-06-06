package main

import (
	"context"
	"time"

	"github.com/ironzhang/x-pearls/zlog"
	"github.com/ironzhang/zerone"
	"github.com/ironzhang/zerone/examples/rpc/arith"
	"github.com/ironzhang/zerone/examples/zerone/conf"
)

func main() {
	defer time.Sleep(10 * time.Millisecond)

	cfg, err := conf.LoadZeroneConfig("../conf/cfg.json")
	if err != nil {
		zlog.Fatalf("load zerone config: %v", err)
	}

	opts, err := cfg.ZeroneOptions()
	if err != nil {
		zlog.Fatalf("get zerone options from config: %v", err)
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

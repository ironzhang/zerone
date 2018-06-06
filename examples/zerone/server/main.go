package main

import (
	"os"
	"os/signal"
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

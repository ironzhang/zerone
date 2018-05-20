package etcdgo

import (
	"context"

	"github.com/coreos/etcd/client"
	"github.com/ironzhang/zerone/route"
	"github.com/ironzhang/zerone/route/governance/etcdgo/etcdapi"
)

type Consumer struct {
	api     *etcdapi.API
	dir     string
	refresh func([]route.Endpoint)
}

func NewConsumer(api client.KeysAPI, dir string, refresh func([]route.Endpoint)) *Consumer {
	c := &Consumer{
		api:     etcdapi.NewAPI(api),
		dir:     dir,
		refresh: refresh,
	}
	return c
}

func (c *Consumer) Close() error {
	return nil
}

func (c *Consumer) watching() {
	var (
		err error
		idx uint64
		evt etcdapi.Event
		eps []route.Endpoint
	)
	for {
		eps, idx, err = c.api.Get(context.Background(), c.dir)
		if err == nil {
			break
		}
	}
	c.setup(eps)

	w := c.api.Watcher(c.dir, idx)
	for {
		evt, err = w.Next(context.Background())
		if err != nil {
			continue
		}
		c.update(evt)
	}
}

func (c *Consumer) setup(eps []route.Endpoint) {
}

func (c *Consumer) update(evt etcdapi.Event) {
}

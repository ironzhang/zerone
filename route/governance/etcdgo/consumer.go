package etcdgo

import (
	"github.com/coreos/etcd/client"
	"github.com/ironzhang/zerone/route"
)

type Consumer struct {
	api client.KeysAPI
	dir string
}

func NewConsumer(api client.KeysAPI, dir string, refresh func([]route.Endpoint)) *Consumer {
	c := &Consumer{}
	return c
}

func (c *Consumer) Close() error {
	return nil
}

func (c *Consumer) monitor() {
}

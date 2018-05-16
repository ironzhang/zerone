package etcdgo

import (
	"fmt"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/ironzhang/zerone/route"
)

type Client struct {
	api       client.KeysAPI
	namespace string
}

func (c *Client) NewProvider(service string, endpoint func() route.Endpoint) *Provider {
	return NewProvider(c.api, fmt.Sprintf("%s/%s", c.namespace, service), 10*time.Second, endpoint)
}

func (c *Client) NewConsumer(service string, refresh func([]route.Endpoint)) {
}

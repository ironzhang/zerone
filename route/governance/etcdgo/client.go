package etcdgo

import (
	"context"
	"fmt"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/ironzhang/zerone/route"
)

type Client struct {
	namespace string
	client    client.Client
	api       client.KeysAPI
}

func NewClient(namespace string, config client.Config) (*Client, error) {
	c, err := client.New(config)
	if err != nil {
		return nil, err
	}
	if err = c.Sync(context.Background()); err != nil {
		return nil, err
	}
	return &Client{
		namespace: namespace,
		client:    c,
		api:       client.NewKeysAPI(c),
	}, nil
}

func (c *Client) NewProvider(service string, endpoint func() route.Endpoint) *Provider {
	return NewProvider(c.api, c.dir(service), 10*time.Second, endpoint)
}

func (c *Client) NewConsumer(service string, refresh func([]route.Endpoint)) *Consumer {
	return NewConsumer(c.api, c.dir(service), refresh)
}

func (c *Client) dir(service string) string {
	return fmt.Sprintf("/%s/%s", c.namespace, service)
}

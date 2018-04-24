package zerone

import (
	"context"

	"github.com/ironzhang/zerone/route"
	"github.com/ironzhang/zerone/route/balance"
)

type LoadBalancer string

const (
	HashBalancer       LoadBalancer = "HashBalancer"
	RandomBalancer     LoadBalancer = "RandomBalancer"
	RoundRobinBalancer LoadBalancer = "RoundRobinBalancer"
)

type Client struct {
	table              route.Table
	balancer           route.LoadBalancer
	hashBalancer       *balance.HashBalancer
	randomBalancer     *balance.RandomBalancer
	roundRobinBalancer *balance.RoundRobinBalancer
}

func NewClient(table route.Table) *Client {
	c := &Client{
		table:              table,
		hashBalancer:       balance.NewHashBalancer(table, nil),
		randomBalancer:     balance.NewRandomBalancer(table),
		roundRobinBalancer: balance.NewRoundRobinBalancer(table),
	}
	c.balancer = c.randomBalancer
	return c
}

func (c *Client) Close() error {
	return nil
}

func (c *Client) getLoadBalancer(lb LoadBalancer) route.LoadBalancer {
	switch lb {
	case HashBalancer:
		return c.hashBalancer
	case RandomBalancer:
		return c.randomBalancer
	case RoundRobinBalancer:
		return c.roundRobinBalancer
	default:
		return c.randomBalancer
	}
}

func (c *Client) WithLoadBalancer(lb LoadBalancer) *Client {
	return &Client{
		table:              c.table,
		balancer:           c.getLoadBalancer(lb),
		hashBalancer:       c.hashBalancer,
		randomBalancer:     c.randomBalancer,
		roundRobinBalancer: c.roundRobinBalancer,
	}
}

func (c *Client) WithFailPolicy(fp FailPolicy) *Client {
	return &Client{}
}

func (c *Client) Call(ctx context.Context, method string, args, res interface{}) error {
	return nil
}

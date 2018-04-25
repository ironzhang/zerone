package zerone

import (
	"context"
	"sync"

	"github.com/ironzhang/zerone/route"
	"github.com/ironzhang/zerone/route/balance"
	"github.com/ironzhang/zerone/rpc"
)

type LoadBalancer string

const (
	HashBalancer       LoadBalancer = "HashBalancer"
	RandomBalancer     LoadBalancer = "RandomBalancer"
	RoundRobinBalancer LoadBalancer = "RoundRobinBalancer"
)

type Client struct {
	name               string
	verbose            int
	table              route.Table
	balancer           route.LoadBalancer
	hashBalancer       *balance.HashBalancer
	randomBalancer     *balance.RandomBalancer
	roundRobinBalancer *balance.RoundRobinBalancer
	clientMap          sync.Map
}

func NewClient(name string, table route.Table) *Client {
	c := &Client{
		name:               name,
		verbose:            0,
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

func (c *Client) SetTraceVerbose(verbose int) {
	c.verbose = verbose
	c.clientMap.Range(func(key, value interface{}) bool {
		rc := value.(*rpc.Client)
		rc.SetTraceVerbose(verbose)
		return true
	})
}

func (c *Client) WithFailPolicy(fp FailPolicy) *Client {
	return &Client{}
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
		verbose:            c.verbose,
		balancer:           c.getLoadBalancer(lb),
		hashBalancer:       c.hashBalancer,
		randomBalancer:     c.randomBalancer,
		roundRobinBalancer: c.roundRobinBalancer,
	}
}

func (c *Client) dial(addr string) (*rpc.Client, error) {
	if value, ok := c.clientMap.Load(addr); ok {
		return value.(*rpc.Client), nil
	}

	client, err := rpc.Dial(c.name, "tcp", addr)
	if err != nil {
		return nil, err
	}
	client.SetTraceVerbose(c.verbose)

	if value, ok := c.clientMap.LoadOrStore(addr, client); ok {
		client.Close()
		return value.(*rpc.Client), nil
	}
	return client, nil
}

func (c *Client) Call(ctx context.Context, method string, key []byte, args, res interface{}) error {
	ep, err := c.balancer.GetEndpoint(key)
	if err != nil {
		return err
	}
	rc, err := c.dial(ep.Addr)
	if err != nil {
		return err
	}
	return rc.Call(ctx, method, args, res)
}

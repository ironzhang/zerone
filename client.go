package zerone

import (
	"context"
	"fmt"
	"io"
	"sync/atomic"

	"github.com/ironzhang/zerone/route"
	"github.com/ironzhang/zerone/route/balance"
	"github.com/ironzhang/zerone/rpc"
)

type LoadBalancer string

// 负载均衡策略定义
const (
	HashBalancer       LoadBalancer = "HashBalancer"
	RandomBalancer     LoadBalancer = "RandomBalancer"
	RoundRobinBalancer LoadBalancer = "RoundRobinBalancer"
)

type FailPolicy string

// 失败重试策略定义
const (
	Failfast FailPolicy = "Failfast"
	Failover FailPolicy = "Failover"
	Failtry  FailPolicy = "Failtry"
)

type Client struct {
	shutdown           int32
	clientset          *clientset
	table              route.Table
	balancer           route.LoadBalancer
	hashBalancer       *balance.HashBalancer
	randomBalancer     *balance.RandomBalancer
	roundRobinBalancer *balance.RoundRobinBalancer
	failPolicy         FailPolicy
}

func NewClient(name string, table route.Table) *Client {
	c := &Client{
		shutdown:           0,
		clientset:          newClientset(name, nil, 0),
		table:              table,
		hashBalancer:       balance.NewHashBalancer(table, nil),
		randomBalancer:     balance.NewRandomBalancer(table),
		roundRobinBalancer: balance.NewRoundRobinBalancer(table),
		failPolicy:         Failfast,
	}
	c.balancer = c.randomBalancer
	return c
}

func (c *Client) clone() *Client {
	return &Client{
		shutdown:           c.shutdown,
		clientset:          c.clientset,
		table:              c.table,
		balancer:           c.balancer,
		hashBalancer:       c.hashBalancer,
		randomBalancer:     c.randomBalancer,
		roundRobinBalancer: c.roundRobinBalancer,
		failPolicy:         c.failPolicy,
	}
}

func (c *Client) Close() error {
	if atomic.CompareAndSwapInt32(&c.shutdown, 0, 1) {
		c.clientset.close()
		return nil
	}
	return rpc.ErrShutdown
}

func (c *Client) SetTraceOutput(output io.Writer) {
	c.clientset.setTraceOutput(output)
}

func (c *Client) SetTraceVerbose(verbose int) {
	c.clientset.setTraceVerbose(verbose)
}

func (c *Client) WithFailPolicy(fp FailPolicy) *Client {
	nc := c.clone()
	nc.failPolicy = fp
	return nc
}

func (c *Client) WithLoadBalancer(lb LoadBalancer) *Client {
	nc := c.clone()
	nc.balancer = nc.getLoadBalancer(lb)
	return nc
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

func (c *Client) Call(ctx context.Context, method string, key []byte, args, res interface{}) error {
	if atomic.LoadInt32(&c.shutdown) == 1 {
		return rpc.ErrShutdown
	}

	for i := 0; i < 3; i++ {
		ep, err := c.balancer.GetEndpoint(key)
		if err != nil {
			return err
		}
		key := fmt.Sprintf("%s://%s", ep.Net, ep.Addr)
		rc, err := c.clientset.add(key, ep.Net, ep.Addr)
		if err != nil {
			continue
		}
		if err = rc.Call(ctx, method, args, res); err == rpc.ErrUnavailable {
			c.clientset.remove(key)
			continue
		}
		return err
	}
	return rpc.ErrUnavailable
}

func (c *Client) failfastCall(ctx context.Context, method string, key []byte, args, res interface{}) error {
	ep, err := c.balancer.GetEndpoint(key)
	if err != nil {
		return err
	}
	target := fmt.Sprintf("%s://%s", ep.Net, ep.Addr)
	rc, err := c.clientset.add(target, ep.Net, ep.Addr)
	if err != nil {
		return err
	}
	if err = rc.Call(ctx, method, args, res); err == rpc.ErrUnavailable {
		c.clientset.remove(target)
	}
	return err
}

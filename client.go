package zerone

import (
	"context"
	"fmt"
	"io"
	"sync/atomic"

	"github.com/ironzhang/zerone/route"
	"github.com/ironzhang/zerone/rpc"
)

type FailPolicy string

// 失败重试策略定义
const (
	Failfast FailPolicy = "Failfast"
	Failover FailPolicy = "Failover"
	Failtry  FailPolicy = "Failtry"
)

type Client struct {
	shutdown      int32
	clientset     *clientset
	balancerset   *balancerset
	balancePolicy BalancePolicy
	failPolicy    FailPolicy
}

func NewClient(name string, table route.Table) *Client {
	return &Client{
		shutdown:      0,
		clientset:     newClientset(name, nil, 0),
		balancerset:   newBalancerset(table),
		balancePolicy: RandomBalancer,
		failPolicy:    Failfast,
	}
}

func (c *Client) clone() *Client {
	return &Client{
		shutdown:      c.shutdown,
		clientset:     c.clientset,
		balancerset:   c.balancerset,
		balancePolicy: c.balancePolicy,
		failPolicy:    c.failPolicy,
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

func (c *Client) WithBalancePolicy(policy BalancePolicy) *Client {
	nc := c.clone()
	nc.balancePolicy = policy
	return nc
}

func (c *Client) WithFailPolicy(policy FailPolicy) *Client {
	nc := c.clone()
	nc.failPolicy = policy
	return nc
}

func (c *Client) Call(ctx context.Context, method string, key []byte, args, res interface{}) error {
	if atomic.LoadInt32(&c.shutdown) == 1 {
		return rpc.ErrShutdown
	}

	for i := 0; i < 3; i++ {
		ep, err := c.balancerset.getLoadBalancer(c.balancePolicy).GetEndpoint(key)
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
	ep, err := c.balancerset.getLoadBalancer(c.balancePolicy).GetEndpoint(key)
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

package zerone

import (
	"context"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/ironzhang/zerone/route"
	"github.com/ironzhang/zerone/rpc"
)

type Client struct {
	shutdown      *int32
	table         route.Table
	clientset     *clientset
	balancerset   *balancerset
	balancePolicy BalancePolicy
	failPolicy    FailPolicy
}

func NewClient(name string, table route.Table) *Client {
	var shutdown int32
	return &Client{
		shutdown:      &shutdown,
		table:         table,
		clientset:     newClientset(name, nil, 0),
		balancerset:   newBalancerset(table),
		balancePolicy: RandomBalancer,
		failPolicy:    NewFailtry(0, 0, 0),
	}
}

func (c *Client) clone() *Client {
	return &Client{
		shutdown:      c.shutdown,
		table:         c.table,
		clientset:     c.clientset,
		balancerset:   c.balancerset,
		balancePolicy: c.balancePolicy,
		failPolicy:    c.failPolicy,
	}
}

func (c *Client) Close() error {
	if atomic.CompareAndSwapInt32(c.shutdown, 0, 1) {
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

func (c *Client) Go(ctx context.Context, key []byte, method string, args, res interface{}, done chan *rpc.Call) (*rpc.Call, error) {
	if atomic.LoadInt32(c.shutdown) == 1 {
		return nil, rpc.ErrShutdown
	}

	lb := c.balancerset.getLoadBalancer(c.balancePolicy)
	return c.failPolicy.execute(lb, key, func(net, addr string) (*rpc.Call, error) {
		rc, err := c.clientset.dial(fmt.Sprintf("%s://%s", net, addr), net, addr)
		if err != nil {
			return nil, err
		}
		return rc.Go(ctx, method, args, res, done)
	})
}

func (c *Client) Call(ctx context.Context, key []byte, method string, args, res interface{}, timeout time.Duration) error {
	call, err := c.Go(ctx, key, method, args, res, make(chan *rpc.Call, 1))
	if err != nil {
		return err
	}

	var tc <-chan time.Time
	if timeout > 0 {
		t := time.NewTimer(timeout)
		defer t.Stop()
		tc = t.C
	}
	select {
	case <-call.Done:
		return call.Error
	case <-tc:
		return rpc.ErrTimeout
	}
}

func (c *Client) Broadcast(ctx context.Context, method string, args, res interface{}) chan *rpc.Call {
	//	eps := c.table.ListEndpoints()
	//	done := make(*rpc.Call, len(eps))
	//	for _, ep := range eps {
	//		rc, err := c.clientset.dial(fmt.Sprintf("%s://%s", net, addr), net, addr)
	//		if err != nil {
	//			continue
	//		}
	//		rc.Go(ctx, method, args, res, done)
	//	}
	//	return done
	return nil
}

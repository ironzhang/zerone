package zclient

import (
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ironzhang/zerone/pkg/endpoint"
	"github.com/ironzhang/zerone/pkg/route"
	"github.com/ironzhang/zerone/rpc"
)

type Client struct {
	shutdown      *int32
	table         route.Table
	connector     *connector
	balancerset   *balancerset
	balancePolicy BalancePolicy
	failPolicy    FailPolicy
}

func New(name string, table route.Table) *Client {
	return &Client{
		shutdown:      new(int32),
		table:         table,
		connector:     newConnector(name, nil, 0),
		balancerset:   newBalancerset(table),
		balancePolicy: RandomBalancer,
		failPolicy:    NewFailtry(0, 0, 0),
	}
}

func (c *Client) clone() *Client {
	return &Client{
		shutdown:      c.shutdown,
		table:         c.table,
		connector:     c.connector,
		balancerset:   c.balancerset,
		balancePolicy: c.balancePolicy,
		failPolicy:    c.failPolicy,
	}
}

func (c *Client) Close() error {
	if atomic.CompareAndSwapInt32(c.shutdown, 0, 1) {
		c.connector.close()
		if closer, ok := c.table.(io.Closer); ok {
			closer.Close()
		}
		return nil
	}
	return rpc.ErrShutdown
}

func (c *Client) SetTraceOutput(output io.Writer) {
	c.connector.setTraceOutput(output)
}

func (c *Client) SetTraceVerbose(verbose int) {
	c.connector.setTraceVerbose(verbose)
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
		rc, err := c.connector.dial(fmt.Sprintf("%s://%s", net, addr), net, addr)
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

type Result struct {
	Endpoint endpoint.Endpoint
	Error    error
	Method   string
	Args     interface{}
	Reply    interface{}
}

func (c *Client) Broadcast(ctx context.Context, method string, args, res interface{}, timeout time.Duration) <-chan Result {
	var wg sync.WaitGroup
	eps := c.table.ListEndpoints()
	ch := make(chan Result, len(eps))
	for _, ep := range eps {
		rc, err := c.connector.dial(fmt.Sprintf("%s://%s", ep.Net, ep.Addr), ep.Net, ep.Addr)
		if err != nil {
			ch <- Result{
				Endpoint: ep,
				Error:    err,
				Method:   method,
				Args:     args,
			}
			continue
		}
		call, err := rc.Go(ctx, method, args, newValuePtr(res), make(chan *rpc.Call, 1))
		if err != nil {
			ch <- Result{
				Endpoint: ep,
				Error:    err,
				Method:   method,
				Args:     args,
			}
			continue
		}

		wg.Add(1)
		go func(ep endpoint.Endpoint, call *rpc.Call) {
			defer wg.Done()
			var tc <-chan time.Time
			if timeout > 0 {
				t := time.NewTimer(timeout)
				defer t.Stop()
				tc = t.C
			}
			select {
			case <-call.Done:
				ch <- Result{
					Endpoint: ep,
					Error:    call.Error,
					Method:   method,
					Args:     args,
					Reply:    call.Reply,
				}

			case <-tc:
				ch <- Result{
					Endpoint: ep,
					Error:    rpc.ErrTimeout,
					Method:   method,
					Args:     args,
				}
			}
		}(ep, call)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()
	return ch
}
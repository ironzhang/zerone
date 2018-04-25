package zerone

import (
	"context"
	"fmt"
	"time"

	"github.com/ironzhang/zerone/rpc"
)

type FailPolicy interface {
	Call(c *Client, ctx context.Context, method string, key []byte, args, res interface{}) error
}

type Failtry struct {
	try int
	min time.Duration
	max time.Duration
}

func NewFailtry(try int, min, max time.Duration) *Failtry {
	if try <= 0 {
		try = 1
	}
	if min <= 0 {
		min = 100 * time.Millisecond
	}
	if max <= 0 {
		max = time.Second
	}
	return &Failtry{
		try: try,
		min: min,
		max: max,
	}
}

func (p *Failtry) Call(c *Client, ctx context.Context, method string, key []byte, args, res interface{}) error {
	ep, err := c.balancerset.getLoadBalancer(c.balancePolicy).GetEndpoint(key)
	if err != nil {
		return err
	}

	delay := p.min
	token := makeToken(ep.Net, ep.Addr)
	for i := 0; i < p.try; i++ {
		if i > 0 {
			time.Sleep(delay)
			delay *= 2
			if delay > p.max {
				delay = p.max
			}
		}
		rc, err := c.clientset.add(token, ep.Net, ep.Addr)
		if err != nil {
			continue
		}
		if err = rc.Call(ctx, method, args, res); err == rpc.ErrUnavailable {
			c.clientset.remove(token)
			continue
		}
		return err
	}
	return rpc.ErrUnavailable
}

type Failover struct {
	try int
}

func NewFailover(try int) *Failover {
	if try <= 0 {
		try = 1
	}
	return &Failover{
		try: try,
	}
}

func (p *Failover) Call(c *Client, ctx context.Context, method string, key []byte, args, res interface{}) error {
	lb := c.balancerset.getLoadBalancer(c.balancePolicy)
	for i := 0; i < p.try; i++ {
		ep, err := lb.GetEndpoint(key)
		if err != nil {
			return err
		}
		token := makeToken(ep.Net, ep.Addr)
		rc, err := c.clientset.add(token, ep.Net, ep.Addr)
		if err != nil {
			continue
		}
		if err = rc.Call(ctx, method, args, res); err == rpc.ErrUnavailable {
			c.clientset.remove(token)
			continue
		}
		return err
	}
	return rpc.ErrUnavailable
}

func makeToken(net, addr string) string {
	return fmt.Sprintf("%s://%s", net, addr)
}

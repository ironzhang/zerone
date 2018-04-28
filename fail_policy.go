package zerone

import (
	"context"
	"fmt"
	"time"

	"github.com/ironzhang/zerone/route"
	"github.com/ironzhang/zerone/rpc"
)

type FailPolicy interface {
	gocall(ctx context.Context, cs *clientset, lb route.LoadBalancer, key []byte, method string, args, res interface{}, done chan *rpc.Call) (*rpc.Call, error)
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

func (p *Failtry) gocall(ctx context.Context, cs *clientset, lb route.LoadBalancer, key []byte,
	method string, args, res interface{}, done chan *rpc.Call) (*rpc.Call, error) {
	ep, err := lb.GetEndpoint(key)
	if err != nil {
		return nil, err
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
		rc, err := cs.dial(token, ep.Net, ep.Addr)
		if err == rpc.ErrShutdown {
			return nil, err
		} else if err != nil {
			continue
		}
		call, err := rc.Go(ctx, method, args, res, done)
		if err == rpc.ErrShutdown {
			return nil, err
		} else if err != nil {
			continue
		}
		return call, err
	}
	return nil, rpc.ErrUnavailable
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

func (p *Failover) Go(ctx context.Context, cs *clientset, lb route.LoadBalancer, key []byte,
	method string, args, res interface{}, done chan *rpc.Call) (*rpc.Call, error) {
	for i := 0; i < p.try; i++ {
		ep, err := lb.GetEndpoint(key)
		if err != nil {
			return nil, err
		}
		token := makeToken(ep.Net, ep.Addr)
		rc, err := cs.dial(token, ep.Net, ep.Addr)
		if err == rpc.ErrShutdown {
			return nil, err
		} else if err != nil {
			continue
		}
		call, err := rc.Go(ctx, method, args, res, done)
		if err == rpc.ErrShutdown {
			return nil, err
		} else if err != nil {
			continue
		}
		return call, err
	}
	return nil, rpc.ErrUnavailable
}

func makeToken(net, addr string) string {
	return fmt.Sprintf("%s://%s", net, addr)
}

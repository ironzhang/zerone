package zclient

import (
	"time"

	"github.com/ironzhang/zerone/route"
	"github.com/ironzhang/zerone/rpc"
)

var timeSleep = time.Sleep

type FailPolicy interface {
	execute(lb route.LoadBalancer, key []byte, do func(net, addr string) (*rpc.Call, error)) (*rpc.Call, error)
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
	if min > max {
		min = max
	}
	return &Failtry{
		try: try,
		min: min,
		max: max,
	}
}

func (p *Failtry) execute(lb route.LoadBalancer, key []byte, do func(net, addr string) (*rpc.Call, error)) (*rpc.Call, error) {
	ep, err := lb.GetEndpoint(key)
	if err != nil {
		return nil, err
	}

	delay := p.min
	for i := 0; i < p.try; i++ {
		if i > 0 {
			timeSleep(delay)
			delay *= 2
			if delay > p.max {
				delay = p.max
			}
		}
		if call, err := do(ep.Net, ep.Addr); err == rpc.ErrShutdown {
			return nil, err
		} else if err != nil {
			continue
		} else {
			return call, err
		}
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

func (p *Failover) execute(lb route.LoadBalancer, key []byte, do func(net, addr string) (*rpc.Call, error)) (*rpc.Call, error) {
	for i := 0; i < p.try; i++ {
		ep, err := lb.GetEndpoint(key)
		if err != nil {
			return nil, err
		}
		if call, err := do(ep.Net, ep.Addr); err == rpc.ErrShutdown {
			return nil, err
		} else if err != nil {
			continue
		} else {
			return call, err
		}
	}
	return nil, rpc.ErrUnavailable
}

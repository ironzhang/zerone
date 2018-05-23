package etcdv2

import (
	"context"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/ironzhang/zerone/govern"
	"github.com/ironzhang/zerone/govern/etcdv2/etcdapi"
	"github.com/ironzhang/zerone/zlog"
)

type provider struct {
	api      etcdapi.API
	dir      string
	ttl      time.Duration
	interval time.Duration
	endpoint govern.GetEndpointFunc
	done     chan struct{}
}

func newProvider(api client.KeysAPI, dir string, interval time.Duration, endpoint govern.GetEndpointFunc) *provider {
	return new(provider).init(api, dir, interval, endpoint)
}

func (p *provider) init(api client.KeysAPI, dir string, interval time.Duration, endpoint govern.GetEndpointFunc) *provider {
	if endpoint == nil {
		panic("govern.GetEndpointFunc is nil")
	}

	p.api.Init(api, endpoint())
	p.dir = dir
	p.ttl = interval * 3
	p.interval = interval
	p.endpoint = endpoint
	p.done = make(chan struct{})
	go p.pinging(p.done)
	return p
}

func (p *provider) Driver() string {
	return DriverName
}

func (p *provider) Directory() string {
	return p.dir
}

func (p *provider) Close() error {
	close(p.done)
	return nil
}

func (p *provider) pinging(done <-chan struct{}) {
	t := time.NewTicker(p.interval)
	defer t.Stop()

	p.register()
	for {
		select {
		case <-t.C:
			p.update()
		case <-done:
			p.unregister()
			return
		}
	}
}

func (p *provider) register() error {
	ep := p.endpoint()
	if err := p.api.Set(context.Background(), p.dir, ep, p.ttl); err != nil {
		zlog.Warnw("register endpoint", "dir", p.dir, "endpoint", ep, "error", err)
		return err
	}
	zlog.Debugw("register endpoint", "dir", p.dir, "endpoint", ep)
	return nil
}

func (p *provider) unregister() error {
	ep := p.endpoint()
	if err := p.api.Del(context.Background(), p.dir, ep.Node()); err != nil {
		zlog.Warnw("unregister endpoint", "dir", p.dir, "endpoint", ep, "error", err)
		return err
	}
	zlog.Debugw("unregister endpoint", "dir", p.dir, "endpoint", ep)
	return nil
}

func (p *provider) update() error {
	ep := p.endpoint()
	if err := p.api.Set(context.Background(), p.dir, ep, p.ttl); err != nil {
		zlog.Warnw("update endpoint", "dir", p.dir, "endpoint", ep, "error", err)
		return err
	}
	zlog.Debugw("update endpoint", "dir", p.dir, "endpoint", ep)
	return nil
}

package etcdgo

import (
	"context"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/ironzhang/zerone/route"
	"github.com/ironzhang/zerone/route/governance/etcdgo/etcdapi"
	"github.com/ironzhang/zerone/zlog"
)

type Provider struct {
	api      *etcdapi.API
	dir      string
	ttl      time.Duration
	interval time.Duration
	endpoint func() route.Endpoint
	done     chan struct{}
}

func NewProvider(api client.KeysAPI, dir string, interval time.Duration, endpoint func() route.Endpoint) *Provider {
	if endpoint == nil {
		panic("endpoint is nil")
	}

	p := &Provider{
		api:      etcdapi.NewAPI(api),
		dir:      dir,
		ttl:      interval * 3,
		interval: interval,
		endpoint: endpoint,
		done:     make(chan struct{}),
	}
	go p.pinging(p.done)
	return p
}

func (p *Provider) Close() error {
	close(p.done)
	return nil
}

func (p *Provider) pinging(done <-chan struct{}) {
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

func (p *Provider) register() error {
	ep := p.endpoint()
	if err := p.api.Set(context.Background(), p.dir, ep, p.ttl); err != nil {
		zlog.Warnw("register endpoint", "dir", p.dir, "endpoint", ep, "error", err)
		return err
	}
	zlog.Debugw("register endpoint", "dir", p.dir, "endpoint", ep)
	return nil
}

func (p *Provider) unregister() error {
	ep := p.endpoint()
	if err := p.api.Del(context.Background(), p.dir, ep.Name); err != nil {
		zlog.Warnw("unregister endpoint", "dir", p.dir, "endpoint", ep, "error", err)
		return err
	}
	zlog.Debugw("unregister endpoint", "dir", p.dir, "endpoint", ep)
	return nil
}

func (p *Provider) update() error {
	ep := p.endpoint()
	if err := p.api.Set(context.Background(), p.dir, ep, p.ttl); err != nil {
		zlog.Warnw("update endpoint", "dir", p.dir, "endpoint", ep, "error", err)
		return err
	}
	zlog.Debugw("update endpoint", "dir", p.dir, "endpoint", ep)
	return nil
}

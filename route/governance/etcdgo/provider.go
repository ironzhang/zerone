package etcdgo

import (
	"context"
	"encoding/json"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/ironzhang/zerone/route"
	"github.com/ironzhang/zerone/zlog"
)

type kAPI struct {
	api client.KeysAPI
	dir string
	ttl time.Duration
}

func (p *kAPI) set(ctx context.Context, ep route.Endpoint) error {
	key := p.dir + "/" + ep.Name
	value, err := json.Marshal(ep)
	if err != nil {
		return err
	}
	opts := client.SetOptions{TTL: p.ttl}
	_, err = p.api.Set(ctx, key, string(value), &opts)
	return err
}

func (p *kAPI) del(ctx context.Context, name string) error {
	key := p.dir + "/" + name
	_, err := p.api.Delete(ctx, key, nil)
	return err
}

type Provider struct {
	kAPI     kAPI
	interval time.Duration
	endpoint func() route.Endpoint
	done     chan struct{}
}

func NewProvider(api client.KeysAPI, dir string, interval time.Duration, endpoint func() route.Endpoint) *Provider {
	if endpoint == nil {
		panic("endpoint is nil")
	}

	p := &Provider{
		kAPI: kAPI{
			api: api,
			dir: dir,
			ttl: interval * 3,
		},
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
	if err := p.kAPI.set(context.Background(), ep); err != nil {
		zlog.Warnw("register endpoint", "endpoint", ep, "error", err)
		return err
	}
	zlog.Debugw("register endpoint", "endpoint", ep)
	return nil
}

func (p *Provider) unregister() error {
	ep := p.endpoint()
	if err := p.kAPI.del(context.Background(), ep.Name); err != nil {
		zlog.Warnw("unregister endpoint", "endpoint", ep, "error", err)
		return err
	}
	zlog.Debugw("unregister endpoint", "endpoint", ep)
	return nil
}

func (p *Provider) update() error {
	ep := p.endpoint()
	if err := p.kAPI.set(context.Background(), ep); err != nil {
		zlog.Warnw("update endpoint", "endpoint", ep, "error", err)
		return err
	}
	zlog.Debugw("update endpoint", "endpoint", ep)
	return nil
}

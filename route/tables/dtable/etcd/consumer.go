package etcd

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/ironzhang/zerone/route"
	"github.com/ironzhang/zerone/route/tables/dtable/etcd/etcdapi"
	"github.com/ironzhang/zerone/zlog"
)

type Consumer struct {
	api       etcdapi.API
	dir       string
	refresh   func([]route.Endpoint)
	endpoints map[string]route.Endpoint
	done      chan struct{}

	mu   sync.RWMutex
	list []route.Endpoint
}

func NewConsumer(api client.KeysAPI, dir string, refresh func([]route.Endpoint)) *Consumer {
	return new(Consumer).init(api, dir, refresh)
}

func (c *Consumer) init(api client.KeysAPI, dir string, refresh func([]route.Endpoint)) *Consumer {
	c.api.Init(api)
	c.dir = dir
	c.refresh = refresh
	c.endpoints = make(map[string]route.Endpoint)
	c.done = make(chan struct{})
	go c.watching(c.done)
	return c
}

func (c *Consumer) Close() error {
	close(c.done)
	return nil
}

func (c *Consumer) ListEndpoints() []route.Endpoint {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.list
}

func (c *Consumer) watching(done chan struct{}) {
	zlog.Infow("start watch", "dir", c.dir)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-done
		cancel()
	}()

	endpoints, index, err := c.listEndpoints(ctx)
	if err != nil {
		zlog.Infow("stop watch", "dir", c.dir)
		return
	}
	c.setup(endpoints)

	w := c.api.Watcher(c.dir, index)
	for {
		if evt, err := c.watch(ctx, w); err == nil {
			c.update(evt)
		} else {
			break
		}
	}
	zlog.Infow("stop watch", "dir", c.dir)
}

func (c *Consumer) listEndpoints(ctx context.Context) ([]route.Endpoint, uint64, error) {
	const min, max = time.Second, 60 * time.Second

	delay := min
	for {
		endpoints, index, err := c.api.Get(ctx, c.dir)
		if err == nil {
			return endpoints, index, nil
		} else if e, ok := err.(client.Error); ok && e.Code == client.ErrorCodeKeyNotFound {
			return nil, index, nil
		} else if err == context.Canceled {
			return nil, 0, err
		} else {
			zlog.Warnw("list endpoints", "dir", c.dir, "delay", delay, "error", err)
			time.Sleep(delay)
			if delay = delay * 2; delay > max {
				delay = max
			}
			continue
		}
	}
}

func (c *Consumer) watch(ctx context.Context, w *etcdapi.Watcher) (etcdapi.Event, error) {
	const min, max = 5 * time.Millisecond, time.Second

	delay := min
	for {
		event, err := w.Next(ctx)
		if err == nil {
			return event, nil
		} else if err == context.Canceled {
			return event, err
		} else {
			zlog.Warnw("watch next", "dir", c.dir, "delay", delay, "error", err)
			time.Sleep(delay)
			if delay = delay * 2; delay > max {
				delay = max
			}
			continue
		}
	}
}

func (c *Consumer) setup(endpoints []route.Endpoint) {
	zlog.Debugw("setup", "dir", c.dir, "endpoints", endpoints)
	for _, ep := range endpoints {
		c.endpoints[ep.Name] = ep
	}
	c.doRefresh()
}

func (c *Consumer) update(event etcdapi.Event) {
	zlog.Debugw("update", "dir", c.dir, "event", event)
	switch event.Action {
	case "set", "update":
		if ep, ok := c.endpoints[event.Name]; !ok || ep != event.Endpoint {
			c.endpoints[event.Name] = event.Endpoint
			c.doRefresh()
		}
	case "delete", "expire":
		if _, ok := c.endpoints[event.Name]; ok {
			delete(c.endpoints, event.Name)
			c.doRefresh()
		}
	}
}

func (c *Consumer) doRefresh() {
	endpoints := sortEndpoints(c.endpoints)
	c.mu.Lock()
	c.list = endpoints
	c.mu.Unlock()
	if c.refresh != nil {
		c.refresh(endpoints)
	}
}

func sortEndpoints(m map[string]route.Endpoint) []route.Endpoint {
	s := make([]route.Endpoint, 0, len(m))
	for _, p := range m {
		s = append(s, p)
	}
	sort.Slice(s, func(i, j int) bool { return s[i].Name < s[j].Name })
	return s
}

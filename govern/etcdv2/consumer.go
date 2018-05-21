package etcdv2

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/ironzhang/zerone/govern"
	"github.com/ironzhang/zerone/govern/etcdv2/etcdapi"
	"github.com/ironzhang/zerone/zlog"
)

type consumer struct {
	api       etcdapi.API
	dir       string
	refresh   func([]govern.Endpoint)
	endpoints map[string]govern.Endpoint
	done      chan struct{}

	mu   sync.RWMutex
	list []govern.Endpoint
}

func newConsumer(api client.KeysAPI, dir string, endpoint govern.Endpoint, refresh func([]govern.Endpoint)) *consumer {
	return new(consumer).init(api, dir, endpoint, refresh)
}

func (c *consumer) init(api client.KeysAPI, dir string, endpoint govern.Endpoint, refresh func([]govern.Endpoint)) *consumer {
	c.api.Init(api, endpoint)
	c.dir = dir
	c.refresh = refresh
	c.endpoints = make(map[string]govern.Endpoint)
	c.done = make(chan struct{})
	go c.watching(c.done)
	return c
}

func (c *consumer) Driver() string {
	return DriverName
}

func (c *consumer) Directory() string {
	return c.dir
}

func (c *consumer) Close() error {
	close(c.done)
	return nil
}

func (c *consumer) GetEndpoints() []govern.Endpoint {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.list
}

func (c *consumer) watching(done <-chan struct{}) {
	zlog.Infow("start watch", "dir", c.dir)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-done
		cancel()
	}()

	eps, index, err := c.listEndpoints(ctx)
	if err != nil {
		zlog.Infow("stop watch", "dir", c.dir)
		return
	}
	c.setup(eps)

	w := c.api.Watcher(c.dir, index)
	for {
		event, err := c.watchNext(ctx, w)
		if err != nil {
			break
		}
		c.update(event)
	}

	zlog.Infow("stop watch", "dir", c.dir)
}

func (c *consumer) listEndpoints(ctx context.Context) ([]govern.Endpoint, uint64, error) {
	const min, max = time.Second, 60 * time.Second

	delay := min
	for {
		eps, index, err := c.api.Get(ctx, c.dir)
		if err == nil {
			return eps, index, nil
		} else if e, ok := err.(client.Error); ok && e.Code == client.ErrorCodeKeyNotFound {
			return nil, index, nil
		} else if err == context.Canceled {
			return nil, 0, err
		} else {
			zlog.Warnw("list endpoints", "dir", c.dir, "delay", delay, "error", err)
			time.Sleep(delay)
			if delay *= 2; delay > max {
				delay = max
			}
		}
	}
}

func (c *consumer) watchNext(ctx context.Context, w *etcdapi.Watcher) (etcdapi.Event, error) {
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
			if delay *= 2; delay > max {
				delay = max
			}
		}
	}
}

func (c *consumer) setup(eps []govern.Endpoint) {
	zlog.Debugw("setup", "dir", c.dir, "endpoints", eps)
	for _, ep := range eps {
		c.endpoints[ep.Node()] = ep
	}
	c.doRefresh()
}

func (c *consumer) update(event etcdapi.Event) {
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

func (c *consumer) doRefresh() {
	eps := sortEndpoints(c.endpoints)
	c.mu.Lock()
	c.list = eps
	c.mu.Unlock()
	if c.refresh != nil {
		c.refresh(eps)
	}
}

func sortEndpoints(m map[string]govern.Endpoint) []govern.Endpoint {
	s := make([]govern.Endpoint, 0, len(m))
	for _, p := range m {
		s = append(s, p)
	}
	sort.Slice(s, func(i, j int) bool {
		return s[i].Node() < s[j].Node()
	})
	return s
}

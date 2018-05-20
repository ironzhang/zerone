package etcdapi

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/ironzhang/zerone/route"
)

type Event struct {
	Action   string
	Name     string
	Endpoint route.Endpoint
}

type Watcher struct {
	watcher client.Watcher
}

func NewWatcher(w client.Watcher) *Watcher {
	return &Watcher{watcher: w}
}

func parseName(key string) (string, error) {
	i := strings.LastIndex(key, "/")
	if i < 0 {
		return "", fmt.Errorf("invalid key: %s", key)
	}
	return key[i+1:], nil
}

func (p *Watcher) Next(ctx context.Context) (evt Event, err error) {
	res, err := p.watcher.Next(ctx)
	if err != nil {
		return Event{}, err
	}
	name, err := parseName(res.Node.Key)
	if err != nil {
		return Event{}, err
	}
	var ep route.Endpoint
	if err = json.Unmarshal([]byte(res.Node.Value), &ep); err != nil {
		return Event{}, err
	}
	return Event{
		Action:   res.Action,
		Name:     name,
		Endpoint: ep,
	}, nil
}

type API struct {
	api client.KeysAPI
}

func NewAPI(a client.KeysAPI) *API {
	return &API{api: a}
}

func (p *API) Get(ctx context.Context, dir string) (endpoints []route.Endpoint, index uint64, err error) {
	res, err := p.api.Get(ctx, dir, &client.GetOptions{Recursive: true})
	if err != nil {
		return nil, 0, err
	}
	var ep route.Endpoint
	for _, node := range res.Node.Nodes {
		if err = json.Unmarshal([]byte(node.Value), &ep); err != nil {
			return nil, 0, err
		}
		endpoints = append(endpoints, ep)
	}
	return endpoints, res.Index, nil
}

func (p *API) Set(ctx context.Context, dir string, ep route.Endpoint, ttl time.Duration) error {
	key := dir + "/" + ep.Name
	value, err := json.Marshal(ep)
	if err != nil {
		return err
	}
	_, err = p.api.Set(ctx, key, string(value), &client.SetOptions{TTL: ttl})
	return err
}

func (p *API) Del(ctx context.Context, dir, name string) error {
	key := dir + "/" + name
	_, err := p.api.Delete(ctx, key, nil)
	return err
}

func (p *API) Watcher(dir string, index uint64) *Watcher {
	return NewWatcher(p.api.Watcher(dir, &client.WatcherOptions{Recursive: true, AfterIndex: index}))
}

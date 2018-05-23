package etcdapi

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/ironzhang/zerone/govern"
)

type API struct {
	api client.KeysAPI
	typ reflect.Type
	ptr bool
}

func NewAPI(a client.KeysAPI, ep govern.Endpoint) *API {
	return new(API).Init(a, ep)
}

func (p *API) Init(api client.KeysAPI, ep govern.Endpoint) *API {
	typ := reflect.TypeOf(ep)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	p.api = api
	p.typ = typ
	return p
}

func (p *API) Get(ctx context.Context, dir string) (endpoints []govern.Endpoint, index uint64, err error) {
	res, err := p.api.Get(ctx, dir, &client.GetOptions{Recursive: true})
	if err != nil {
		return nil, 0, err
	}
	for _, node := range res.Node.Nodes {
		ep := reflect.New(p.typ).Interface().(govern.Endpoint)
		if err = json.Unmarshal([]byte(node.Value), ep); err != nil {
			return nil, 0, err
		}
		endpoints = append(endpoints, ep)
	}
	return endpoints, res.Index, nil
}

func (p *API) Set(ctx context.Context, dir string, ep govern.Endpoint, ttl time.Duration) error {
	key := dir + "/" + ep.Node()
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
	return &Watcher{
		watcher: p.api.Watcher(dir, &client.WatcherOptions{Recursive: true, AfterIndex: index}),
		typ:     p.typ,
	}
}

type Event struct {
	Action   string
	Name     string
	Endpoint govern.Endpoint
}

type Watcher struct {
	watcher client.Watcher
	typ     reflect.Type
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

	var ep govern.Endpoint
	if res.Action == "set" || res.Action == "update" {
		ep = reflect.New(p.typ).Interface().(govern.Endpoint)
		if err = json.Unmarshal([]byte(res.Node.Value), ep); err != nil {
			return Event{}, err
		}
	}
	return Event{
		Action:   res.Action,
		Name:     name,
		Endpoint: ep,
	}, nil
}

func parseName(key string) (string, error) {
	i := strings.LastIndex(key, "/")
	if i < 0 {
		return "", fmt.Errorf("invalid key: %s", key)
	}
	return key[i+1:], nil
}

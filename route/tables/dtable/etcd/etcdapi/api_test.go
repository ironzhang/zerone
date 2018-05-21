package etcdapi

import (
	"context"
	"testing"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/ironzhang/zerone/route"
)

func TestParseName(t *testing.T) {
	tests := []struct {
		key  string
		name string
	}{
		{
			key:  "/",
			name: "",
		},
		{
			key:  "/foo",
			name: "foo",
		},
		{
			key:  "/foo/bar",
			name: "bar",
		},
		{
			key:  "/foo/",
			name: "",
		},
	}
	for _, tt := range tests {
		name, err := parseName(tt.key)
		if err != nil {
			t.Fatalf("%q: parse name: %v", tt.key, err)
		}
		if got, want := name, tt.name; got != want {
			t.Errorf("%q: parse name: got %q, want %q", tt.key, got, want)
		} else {
			t.Logf("%q parse name: got %q", tt.key, got)
		}
	}
}

func NewEtcdClient(addrs []string) (client.Client, error) {
	return client.New(client.Config{Endpoints: addrs})
}

func NewTestKeysAPI() client.KeysAPI {
	c, err := NewEtcdClient([]string{"http://127.0.0.1:2379"})
	if err != nil {
		panic(err)
	}
	return client.NewKeysAPI(c)
}

func TestAPI(t *testing.T) {
	api := NewAPI(NewTestKeysAPI())

	endpoints := []route.Endpoint{
		{Name: "node1", Net: "tcp", Addr: "localhost:2000"},
		{Name: "node2", Net: "tcp", Addr: "localhost:2000"},
		{Name: "node3", Net: "tcp", Addr: "localhost:2000"},
	}
	for _, ep := range endpoints {
		if err := api.Set(context.Background(), "/TestAPI", ep, 10*time.Second); err != nil {
			t.Fatalf("set: %v", err)
		}
	}
	eps, _, err := api.Get(context.Background(), "/TestAPI")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got, want := len(eps), len(endpoints); got != want {
		t.Fatalf("after set, count: got %v, want %v", got, want)
	}
	t.Logf("after set, endpoints: got %v", eps)

	for _, ep := range endpoints {
		if err := api.Del(context.Background(), "/TestAPI", ep.Name); err != nil {
			t.Fatalf("del: %v", err)
		}
	}
	eps, _, err = api.Get(context.Background(), "/TestAPI")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got, want := len(eps), 0; got != want {
		t.Fatalf("after del, count: got %v, want %v", got, want)
	}
	t.Logf("after del, endpoints: got %v", eps)
}

func TestWatcher(t *testing.T) {
	api := NewAPI(NewTestKeysAPI())

	go func() {
		ep := route.Endpoint{Name: "node1", Net: "tcp", Addr: "localhost:2000"}
		api.Set(context.Background(), "/TestWatcher", ep, 10*time.Second)
		api.Set(context.Background(), "/TestWatcher", ep, 10*time.Second)
		api.Del(context.Background(), "/TestWatcher", ep.Name)
	}()

	w := api.Watcher("/TestWatcher", 0)
	for i := 0; i < 3; i++ {
		evt, err := w.Next(context.Background())
		if err != nil {
			t.Fatalf("next: %v", err)
		}
		t.Logf("event: %v", evt)
	}
}

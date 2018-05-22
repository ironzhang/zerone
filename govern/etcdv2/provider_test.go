package etcdv2

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/ironzhang/zerone/govern"
	"github.com/ironzhang/zerone/govern/etcdv2/etcdapi"
)

func EtcdAPIGetEndpoints(eapi *etcdapi.API, dir string, count int) ([]govern.Endpoint, error) {
	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		eps, _, err := eapi.Get(context.Background(), dir)
		if err != nil {
			return eps, err
		}
		if len(eps) != count {
			continue
		}
		sort.Slice(eps, func(i, j int) bool {
			return eps[i].Node() < eps[j].Node()
		})
		return eps, nil
	}
	return nil, errors.New("timeout")
}

func TestProvider(t *testing.T) {
	defer time.Sleep(500 * time.Millisecond)

	dir := "/TestProvider"
	api := NewTestKeysAPI()
	eapi := etcdapi.NewAPI(api, &Endpoint{})
	endpoints := []*Endpoint{
		&Endpoint{Name: "node0", Addr: "localhost:2000"},
		&Endpoint{Name: "node1", Addr: "localhost:2001"},
		&Endpoint{Name: "node2", Addr: "localhost:2002"},
	}
	for _, ep := range endpoints {
		p := newProvider(api, dir, 10*time.Second, ep)
		defer p.Close()
	}

	eps, err := EtcdAPIGetEndpoints(eapi, dir, len(endpoints))
	if err != nil {
		t.Fatalf("EtcdAPIGetEndpoints: %v", err)
	}

	for i, ep := range endpoints {
		if got, want := eps[i].(*Endpoint), ep; *got != *want {
			t.Fatalf("%d: endpoint: got %v, want %v", i, got, want)
		} else {
			t.Logf("%d: endpoint: got %v", i, got)
		}
	}
}

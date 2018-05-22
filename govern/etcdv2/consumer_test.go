package etcdv2

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ironzhang/zerone/govern"
	"github.com/ironzhang/zerone/govern/etcdv2/etcdapi"
)

func ConsumerGetEndpoints(c govern.Consumer, count int) ([]govern.Endpoint, error) {
	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		if eps := c.GetEndpoints(); len(eps) == count {
			return eps, nil
		}
	}
	return nil, errors.New("timeout")
}

func RefreshTableGetEndpoints(rt *RefreshTable, count int) ([]Endpoint, error) {
	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		if eps := rt.ListEndpoints(); len(eps) == count {
			return eps, nil
		}
	}
	return nil, errors.New("timeout")
}

func TestConsumerGetEndpoints(t *testing.T) {
	defer time.Sleep(500 * time.Millisecond)

	dir := "/TestConsumer"
	api := NewTestKeysAPI()
	eapi := etcdapi.NewAPI(api, &Endpoint{})

	c := newConsumer(api, dir, &Endpoint{}, nil)
	defer c.Close()

	endpoints := []*Endpoint{
		&Endpoint{Name: "node0", Addr: "localhost:2000"},
		&Endpoint{Name: "node1", Addr: "localhost:2001"},
		&Endpoint{Name: "node2", Addr: "localhost:2002"},
		&Endpoint{Name: "node3", Addr: "localhost:2003"},
	}
	for _, ep := range endpoints {
		eapi.Del(context.Background(), dir, ep.Name)
		eapi.Set(context.Background(), dir, ep, 10*time.Second)
	}

	eps, err := ConsumerGetEndpoints(c, len(endpoints))
	if err != nil {
		t.Fatalf("ConsumerGetEndpoints: %v", err)
	}

	for i, ep := range endpoints {
		if got, want := eps[i].(*Endpoint), ep; *got != *want {
			t.Fatalf("%d: endpoint: got %v, want %v", i, got, want)
		} else {
			t.Logf("%d: endpoint: got %v", i, got)
		}
	}
}

func TestConsumerRefresh(t *testing.T) {
	defer time.Sleep(500 * time.Millisecond)

	rt := &RefreshTable{}
	dir := "/TestConsumerRefresh"
	api := NewTestKeysAPI()
	eapi := etcdapi.NewAPI(api, &Endpoint{})

	c := newConsumer(api, dir, &Endpoint{}, rt.Refresh)
	defer c.Close()

	endpoints := []*Endpoint{
		&Endpoint{Name: "node0", Addr: "localhost:2000"},
		&Endpoint{Name: "node1", Addr: "localhost:2001"},
		&Endpoint{Name: "node2", Addr: "localhost:2002"},
		&Endpoint{Name: "node3", Addr: "localhost:2003"},
	}
	for _, ep := range endpoints {
		eapi.Del(context.Background(), dir, ep.Name)
		eapi.Set(context.Background(), dir, ep, 10*time.Second)
	}

	eps, err := RefreshTableGetEndpoints(rt, len(endpoints))
	if err != nil {
		t.Fatalf("RefreshTableGetEndpoints: %v", err)
	}

	for i, ep := range endpoints {
		if got, want := &eps[i], ep; *got != *want {
			t.Fatalf("%d: endpoint: got %v, want %v", i, got, want)
		} else {
			t.Logf("%d: endpoint: got %v", i, got)
		}
	}
}

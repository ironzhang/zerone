package etcdv2

import (
	"testing"
	"time"

	"github.com/ironzhang/zerone/govern"
)

func TestDriver(t *testing.T) {
	defer time.Sleep(500 * time.Millisecond)

	ns := "TestDriver"
	sv := "TestService"
	rt := &RefreshTable{}

	d := OpenTestDriver(ns)
	defer d.Close()
	c := d.NewConsumer(sv, &Endpoint{}, rt.Refresh)
	defer c.Close()

	endpoints := []*Endpoint{
		&Endpoint{Name: "node0", Addr: "localhost:2000"},
		&Endpoint{Name: "node1", Addr: "localhost:2001"},
		&Endpoint{Name: "node2", Addr: "localhost:2002"},
		&Endpoint{Name: "node3", Addr: "localhost:2003"},
	}
	for _, ep := range endpoints {
		x := ep
		p := d.NewProvider(sv, 10*time.Second, func() govern.Endpoint { return x })
		defer p.Close()
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

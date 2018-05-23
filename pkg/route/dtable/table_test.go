package dtable

import (
	"errors"
	"testing"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/ironzhang/zerone/govern"
	"github.com/ironzhang/zerone/govern/etcdv2"
	"github.com/ironzhang/zerone/pkg/endpoint"
)

func init() {
	//zlog.Default.SetLevel(zlog.DEBUG)
}

func OpenTestDriver(namespace string) govern.Driver {
	driver, err := govern.Open(etcdv2.DriverName, namespace, client.Config{Endpoints: []string{"http://127.0.0.1:2379"}})
	if err != nil {
		panic(err)
	}
	return driver
}

func ListEndpoints(t *Table, count int) ([]endpoint.Endpoint, error) {
	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		if eps := t.ListEndpoints(); len(eps) == count {
			return eps, nil
		}
	}
	return nil, errors.New("timeout")
}

func TestTable(t *testing.T) {
	defer time.Sleep(100 * time.Millisecond)

	ns := "TestTable"
	sv := "TestService"

	d := OpenTestDriver(ns)
	defer d.Close()

	tb := NewTable(d, sv)
	defer tb.Close()

	endpoints := []*endpoint.Endpoint{
		&endpoint.Endpoint{Name: "node0", Net: "tcp", Addr: "localhost:2000", Load: 0.0},
		&endpoint.Endpoint{Name: "node1", Net: "udp", Addr: "localhost:2001", Load: 0.1},
		&endpoint.Endpoint{Name: "node2", Net: "tcp", Addr: "localhost:2002", Load: 0.2},
		&endpoint.Endpoint{Name: "node3", Net: "udp", Addr: "localhost:2003", Load: 0.3},
	}
	for _, ep := range endpoints {
		x := ep
		p := d.NewProvider(sv, 10*time.Second, func() govern.Endpoint { return x })
		defer p.Close()
	}

	eps, err := ListEndpoints(tb, len(endpoints))
	if err != nil {
		t.Fatalf("ListEndpoints: %v", err)
	}
	for i, ep := range endpoints {
		if got, want := eps[i], *ep; got != want {
			t.Fatalf("%d: endpoint: got %v, want %v", i, got, want)
		} else {
			t.Logf("%d: endpoint: got %v", i, got)
		}
	}
}

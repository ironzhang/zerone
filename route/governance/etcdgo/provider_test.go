package etcdgo

import (
	"context"
	"testing"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/ironzhang/zerone/route"
)

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

func TestKAPI(t *testing.T) {
	api := NewTestKeysAPI()
	kapi := kAPI{api: api, dir: "/test", ttl: 10 * time.Second}

	ep := route.Endpoint{
		Name: "node1",
		Net:  "tcp",
		Addr: "localhost:2000",
		Load: 0.1,
	}
	if err := kapi.set(context.Background(), ep); err != nil {
		t.Fatalf("set: %v", err)
	}
	if err := kapi.del(context.Background(), ep.Name); err != nil {
		t.Fatalf("del: %v", err)
	}
	//	if err := kapi.del(context.Background(), "node1"); err != nil {
	//		t.Fatalf("del: %v", err)
	//	}
}

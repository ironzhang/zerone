package etcdgo

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/ironzhang/zerone/route"
)

func init() {
	//	zlog.Default.SetLevel(zlog.DEBUG)
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

func TestProvider(t *testing.T) {
	api := NewTestKeysAPI()
	p := NewProvider(api, "/TestProvider", 1*time.Second, func() route.Endpoint {
		return route.Endpoint{
			Name: "node1",
			Net:  "tcp",
			Addr: "localhost:2000",
			Load: 0.1,
		}
	})
	time.Sleep(10 * time.Second)
	p.Close()
	time.Sleep(time.Second)
}

func TestConsumer(t *testing.T) {
	api := NewTestKeysAPI()
	c := NewConsumer(api, "/TestConsumer", nil)

	ep := route.Endpoint{
		Name: "node1",
		Net:  "tcp",
		Addr: "localhost:2000",
		Load: 0.1,
	}
	c.api.Set(context.Background(), "/TestConsumer", ep, 10*time.Second)
	c.api.Set(context.Background(), "/TestConsumer", ep, 10*time.Second)
	c.api.Del(context.Background(), "/TestConsumer", ep.Name)
}

func TestClient(t *testing.T) {
	c, err := NewClient("TestClient", client.Config{Endpoints: []string{"http://127.0.0.1:2379"}})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer time.Sleep(time.Second)

	provider := c.NewProvider("TestService", func() route.Endpoint {
		return route.Endpoint{
			Name: "node1",
			Net:  "tcp",
			Addr: "localhost:2000",
			//Load: rand.Float64(),
		}
	})
	defer provider.Close()

	consumer := c.NewConsumer("TestService", func(endpoints []route.Endpoint) {
		log.Printf("endpoints: %v\n", endpoints)
	})
	defer consumer.Close()

	time.Sleep(30 * time.Second)
	log.Printf("last endpoints: %v\n", consumer.GetEndpoints())
}

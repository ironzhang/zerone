package etcdgo

import (
	"testing"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/ironzhang/zerone/route"
	"github.com/ironzhang/zerone/zlog"
)

func init() {
	zlog.Default.SetLevel(zlog.DEBUG)
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
	p := NewProvider(api, "/test", 1*time.Second, func() route.Endpoint {
		return route.Endpoint{
			Name: "node2",
			Net:  "tcp",
			Addr: "localhost:2000",
			Load: 0.1,
		}
	})
	time.Sleep(10 * time.Second)
	p.Close()
	time.Sleep(time.Second)
}

package etcdv2

import (
	"encoding/json"
	"fmt"

	"github.com/coreos/etcd/client"
	"github.com/ironzhang/zerone/govern"
)

func init() {
	//zlog.Default.SetLevel(zlog.DEBUG)
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

func OpenTestDriver(namespace string) govern.Driver {
	driver, err := Open(namespace, client.Config{Endpoints: []string{"http://127.0.0.1:2379"}})
	if err != nil {
		panic(err)
	}
	return driver
}

type Endpoint struct {
	Name string
	Addr string
}

func (p *Endpoint) Node() string {
	return p.Name
}

func (p *Endpoint) String() string {
	return fmt.Sprintf("Name: %s, Addr: %s", p.Name, p.Addr)
}

func (p *Endpoint) Equal(a interface{}) bool {
	ep := a.(*Endpoint)
	return *p == *ep
}

func (p *Endpoint) Marshal() (string, error) {
	data, err := json.Marshal(p)
	return string(data), err
}

func (p *Endpoint) Unmarshal(s string) error {
	return json.Unmarshal([]byte(s), p)
}

type RefreshTable struct {
	list []Endpoint
}

func (t *RefreshTable) Refresh(goeps []govern.Endpoint) {
	list := make([]Endpoint, 0, len(goeps))
	for _, p := range goeps {
		ep := p.(*Endpoint)
		list = append(list, *ep)
	}
	t.list = list
}

func (t *RefreshTable) ListEndpoints() []Endpoint {
	return t.list
}

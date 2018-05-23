package route

import (
	"encoding/json"

	"github.com/ironzhang/zerone/govern"
)

type Endpoint struct {
	Name string
	Net  string
	Addr string
	Load float64
}

func (p *Endpoint) Node() string {
	return p.Name
}

func (p *Endpoint) String() string {
	data, _ := json.Marshal(p)
	return string(data)
}

func (p *Endpoint) Equal(a govern.Endpoint) bool {
	return *p == *a.(*Endpoint)
}

type Table interface {
	ListEndpoints() []Endpoint
}

type LoadBalancer interface {
	GetEndpoint(key []byte) (Endpoint, error)
}

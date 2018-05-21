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

func (p *Endpoint) Marshal() (string, error) {
	data, err := json.Marshal(p)
	return string(data), err
}

func (p *Endpoint) Unmarshal(s string) error {
	return json.Unmarshal([]byte(s), p)
}

func (p *Endpoint) Equal(a govern.Endpoint) bool {
	ep := a.(*Endpoint)
	return *p == *ep
}

type Table interface {
	ListEndpoints() []Endpoint
}

type LoadBalancer interface {
	GetEndpoint(key []byte) (Endpoint, error)
}

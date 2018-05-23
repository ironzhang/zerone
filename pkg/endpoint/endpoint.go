package endpoint

import (
	"encoding/json"
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

func (p *Endpoint) Equal(a interface{}) bool {
	return *p == *a.(*Endpoint)
}

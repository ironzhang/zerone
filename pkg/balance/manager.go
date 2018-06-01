package balance

import (
	"fmt"

	"github.com/ironzhang/zerone/pkg/route"
)

type Manager struct {
	m map[string]LoadBalancer
	d LoadBalancer
}

func NewManager(table route.Table, hash Hash) *Manager {
	return new(Manager).Init(table, hash)
}

func (p *Manager) Init(table route.Table, hash Hash) *Manager {
	p.m = make(map[string]LoadBalancer)
	p.m[RandomBalancerName] = NewRandomBalancer(table)
	p.m[RoundRobinBalancerName] = NewRoundRobinBalancer(table)
	p.m[HashBalancerName] = NewHashBalancer(table, hash)
	p.m[NodeBalancerName] = NewNodeBalancer(table)
	p.d = p.m[RandomBalancerName]
	return p
}

func (p *Manager) GetLoadBalancer(name string) LoadBalancer {
	if lb, ok := p.m[name]; ok {
		return lb
	}
	return p.d
}

func (p *Manager) SetDefaultLoadBalancer(name string) error {
	if lb, ok := p.m[name]; ok {
		p.d = lb
		return nil
	}
	return fmt.Errorf("unknown %q load balancer name", name)
}

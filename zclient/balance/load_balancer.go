package balance

import (
	"errors"
	"hash/crc32"
	"math/rand"

	"github.com/ironzhang/zerone/route"
)

var (
	ErrNoEndpoint = errors.New("no endpoint")
)

type LoadBalancer interface {
	GetEndpoint(key []byte) (route.Endpoint, error)
}

type Hash func(data []byte) uint32

var _ LoadBalancer = &RandomBalancer{}

type RandomBalancer struct {
	table route.Table
}

func NewRandomBalancer(table route.Table) *RandomBalancer {
	return &RandomBalancer{table: table}
}

func (b *RandomBalancer) GetEndpoint(key []byte) (route.Endpoint, error) {
	eps := b.table.ListEndpoints()
	if n := len(eps); n > 0 {
		return eps[rand.Intn(n)], nil
	}
	return route.Endpoint{}, ErrNoEndpoint
}

var _ LoadBalancer = &RoundRobinBalancer{}

type RoundRobinBalancer struct {
	table route.Table
	index uint32
}

func NewRoundRobinBalancer(table route.Table) *RoundRobinBalancer {
	return &RoundRobinBalancer{table: table, index: 0}
}

func (b *RoundRobinBalancer) GetEndpoint(key []byte) (route.Endpoint, error) {
	eps := b.table.ListEndpoints()
	if n := uint32(len(eps)); n > 0 {
		i := b.index % n
		b.index++
		return eps[i], nil
	}
	return route.Endpoint{}, ErrNoEndpoint
}

var _ LoadBalancer = &HashBalancer{}

type HashBalancer struct {
	table route.Table
	hash  Hash
}

func NewHashBalancer(table route.Table, hash Hash) *HashBalancer {
	if hash == nil {
		hash = crc32.ChecksumIEEE
	}
	return &HashBalancer{table: table, hash: hash}
}

func (b *HashBalancer) GetEndpoint(key []byte) (route.Endpoint, error) {
	eps := b.table.ListEndpoints()
	if n := uint32(len(eps)); n > 0 {
		i := b.hash(key) % n
		return eps[i], nil
	}
	return route.Endpoint{}, ErrNoEndpoint
}

var _ LoadBalancer = &NodeBalancer{}

type NodeBalancer struct {
	table route.Table
}

func NewNodeBalancer(table route.Table) *NodeBalancer {
	return &NodeBalancer{table: table}
}

func (b *NodeBalancer) GetEndpoint(key []byte) (route.Endpoint, error) {
	name := string(key)
	eps := b.table.ListEndpoints()
	for _, ep := range eps {
		if ep.Name == name {
			return ep, nil
		}
	}
	return route.Endpoint{}, ErrNoEndpoint
}

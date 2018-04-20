package route

import (
	"errors"
	"hash/crc32"
	"math/rand"
)

var (
	ErrNoEndpoint = errors.New("no endpoint")
)

var (
	_ LoadBalancer = &RandomBalancer{}
	_ LoadBalancer = &RoundRobinBalancer{}
	_ LoadBalancer = &HashBalancer{}
)

type LoadBalancer interface {
	GetEndpoint(key []byte) (Endpoint, error)
}

type RandomBalancer struct {
	table Table
}

func NewRandomBalancer(table Table) *RandomBalancer {
	return &RandomBalancer{table: table}
}

func (b *RandomBalancer) GetEndpoint(key []byte) (Endpoint, error) {
	eps := b.table.ListEndpoints()
	if n := len(eps); n > 0 {
		return eps[rand.Intn(n)], nil
	}
	return Endpoint{}, ErrNoEndpoint
}

type RoundRobinBalancer struct {
	table Table
	index uint32
}

func NewRoundRobinBalancer(table Table) *RoundRobinBalancer {
	return &RoundRobinBalancer{table: table, index: 0}
}

func (b *RoundRobinBalancer) GetEndpoint(key []byte) (Endpoint, error) {
	eps := b.table.ListEndpoints()
	if n := uint32(len(eps)); n > 0 {
		i := b.index % n
		b.index++
		return eps[i], nil
	}
	return Endpoint{}, ErrNoEndpoint
}

type Hash func(data []byte) uint32

type HashBalancer struct {
	table Table
	hash  Hash
}

func NewHashBalancer(table Table, hash Hash) *HashBalancer {
	if hash == nil {
		hash = crc32.ChecksumIEEE
	}
	return &HashBalancer{table: table, hash: hash}
}

func (b *HashBalancer) GetEndpoint(key []byte) (Endpoint, error) {
	eps := b.table.ListEndpoints()
	if n := uint32(len(eps)); n > 0 {
		i := b.hash(key) % n
		return eps[i], nil
	}
	return Endpoint{}, ErrNoEndpoint
}

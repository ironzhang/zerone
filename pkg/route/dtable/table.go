package dtable

import (
	"sync"

	"github.com/ironzhang/x-pearls/govern"
	"github.com/ironzhang/zerone/pkg/endpoint"
	"github.com/ironzhang/zerone/pkg/route"
)

type Driver interface {
	NewConsumer(service string, endpoint govern.Endpoint, f govern.RefreshEndpointsFunc) govern.Consumer
}

var _ route.Table = &Table{}

type Table struct {
	consumer  govern.Consumer
	mu        sync.RWMutex
	endpoints []endpoint.Endpoint
}

func NewTable(driver Driver, service string) *Table {
	return new(Table).init(driver, service)
}

func (t *Table) init(driver Driver, service string) *Table {
	t.consumer = driver.NewConsumer(service, &endpoint.Endpoint{}, t.refresh)
	return t
}

func (t *Table) refresh(goeps []govern.Endpoint) {
	eps := make([]endpoint.Endpoint, 0, len(goeps))
	for _, goep := range goeps {
		ep := goep.(*endpoint.Endpoint)
		eps = append(eps, *ep)
	}

	t.mu.Lock()
	t.endpoints = eps
	t.mu.Unlock()
}

func (t *Table) Close() error {
	return t.consumer.Close()
}

func (t *Table) ListEndpoints() []endpoint.Endpoint {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.endpoints
}

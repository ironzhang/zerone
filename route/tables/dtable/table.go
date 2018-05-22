package dtable

import (
	"sync"

	"github.com/ironzhang/zerone/govern"
	"github.com/ironzhang/zerone/route"
)

type Table struct {
	consumer  govern.Consumer
	mu        sync.RWMutex
	endpoints []route.Endpoint
}

func NewTable(driver govern.Driver, service string) *Table {
	return new(Table).init(driver, service)
}

func (t *Table) init(driver govern.Driver, service string) *Table {
	t.consumer = driver.NewConsumer(service, &route.Endpoint{}, t.refresh)
	return t
}

func (t *Table) refresh(goeps []govern.Endpoint) {
	eps := make([]route.Endpoint, 0, len(goeps))
	for _, goep := range goeps {
		ep := goep.(*route.Endpoint)
		eps = append(eps, *ep)
	}

	t.mu.Lock()
	t.endpoints = eps
	t.mu.Unlock()
}

func (t *Table) Close() error {
	return t.consumer.Close()
}

func (t *Table) ListEndpoints() []route.Endpoint {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.endpoints
}

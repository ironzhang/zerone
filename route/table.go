package route

import (
	"sort"
	"sync"
)

type Endpoint struct {
	Name string
	Addr string
	Load float64
}

type Table struct {
	mu        sync.RWMutex
	endpoints map[string]Endpoint
	list      []Endpoint
}

func NewTable() *Table {
	return &Table{endpoints: make(map[string]Endpoint)}
}

func (t *Table) ListEndpoints() []Endpoint {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.list
}

func (t *Table) AddEndpoints(endpoints ...Endpoint) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(endpoints) > 0 {
		for _, ep := range endpoints {
			t.endpoints[ep.Name] = ep
		}
		t.makeList()
	}
}

func (t *Table) RemoveEndpoints(names ...string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(names) > 0 {
		for _, name := range names {
			delete(t.endpoints, name)
		}
		t.makeList()
	}
}

func (t *Table) makeList() {
	list := make([]Endpoint, 0, len(t.endpoints))
	for _, ep := range t.endpoints {
		list = append(list, ep)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Name < list[j].Name
	})
	t.list = list
}

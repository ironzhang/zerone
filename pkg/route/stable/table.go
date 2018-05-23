package stable

import (
	"fmt"
	"sort"

	"github.com/ironzhang/pearls/config"
	"github.com/ironzhang/zerone/pkg/endpoint"
	"github.com/ironzhang/zerone/pkg/route"
)

var _ route.Table = &Table{}

type Table struct {
	endpoints []endpoint.Endpoint
}

func NewTable(endpoints []endpoint.Endpoint) *Table {
	sort.Slice(endpoints, func(i, j int) bool {
		return endpoints[i].Name < endpoints[j].Name
	})
	return &Table{endpoints: endpoints}
}

func (t *Table) ListEndpoints() []endpoint.Endpoint {
	return t.endpoints
}

type Tables map[string][]endpoint.Endpoint

func LoadTables(filename string) (Tables, error) {
	var tables Tables
	if err := config.LoadFromFile(filename, &tables); err != nil {
		return nil, err
	}
	return tables, nil
}

func (t Tables) Lookup(service string) (*Table, error) {
	endpoints, ok := t[service]
	if !ok {
		return nil, fmt.Errorf("service(%s) not found", service)
	}
	if len(endpoints) <= 0 {
		return nil, fmt.Errorf("service(%s) endpoints is empty", service)
	}
	return NewTable(endpoints), nil
}

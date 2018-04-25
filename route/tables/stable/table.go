package stable

import (
	"fmt"
	"sort"

	"github.com/ironzhang/pearls/config"
	"github.com/ironzhang/zerone/route"
)

var _ route.Table = &Table{}

type Table struct {
	endpoints []route.Endpoint
}

func NewTable(endpoints []route.Endpoint) *Table {
	sort.Slice(endpoints, func(i, j int) bool {
		return endpoints[i].Name < endpoints[j].Name
	})
	return &Table{endpoints: endpoints}
}

func (t *Table) ListEndpoints() []route.Endpoint {
	return t.endpoints
}

func LoadTable(filename string, service string) (*Table, error) {
	var cfg map[string][]route.Endpoint
	if err := config.LoadFromFile(filename, &cfg); err != nil {
		return nil, err
	}
	endpoints, ok := cfg[service]
	if !ok || len(endpoints) <= 0 {
		return nil, fmt.Errorf("not found service: %s", service)
	}
	return NewTable(endpoints), nil
}

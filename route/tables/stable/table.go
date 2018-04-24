package stable

import (
	"sort"

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

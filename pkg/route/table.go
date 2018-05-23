package route

import "github.com/ironzhang/zerone/pkg/endpoint"

type Table interface {
	ListEndpoints() []endpoint.Endpoint
}

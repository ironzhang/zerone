package route

type Endpoint struct {
	Name string
	Net  string
	Addr string
	Load float64
}

type Table interface {
	ListEndpoints() []Endpoint
}

type LoadBalancer interface {
	GetEndpoint(key []byte) (Endpoint, error)
}

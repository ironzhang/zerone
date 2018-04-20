package route

type Table interface {
	ListEndpoints() []Endpoint
}

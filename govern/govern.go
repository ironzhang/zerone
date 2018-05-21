package govern

import (
	"fmt"
	"time"
)

type Endpoint interface {
	Node() string
	String() string
	Marshal() (string, error)
	Unmarshal(s string) error
	Equal(ep Endpoint) bool
}

type Provider interface {
	Driver() string
	Directory() string
	Close() error
}

type Consumer interface {
	Driver() string
	Directory() string
	Close() error
	GetEndpoints() []Endpoint
}

type Driver interface {
	Name() string
	Namespace() string
	NewProvider(service string, endpoint Endpoint, interval time.Duration) Provider
	NewConsumer(service string, endpoint Endpoint, refresh func([]Endpoint)) Consumer
	Close() error
}

type OpenFunc func(namespace string, config interface{}) (Driver, error)

var openfuncs map[string]OpenFunc

func Register(driver string, f OpenFunc) {
	if openfuncs == nil {
		openfuncs = make(map[string]OpenFunc)
	}
	if _, ok := openfuncs[driver]; ok {
		panic(fmt.Errorf("driver(%s) registed", driver))
	}
	openfuncs[driver] = f
}

func Open(driver string, namespace string, config interface{}) (Driver, error) {
	open, ok := openfuncs[driver]
	if !ok {
		return nil, fmt.Errorf("driver(%s) not found", driver)
	}
	return open(namespace, config)
}
